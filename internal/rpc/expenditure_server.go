package rpc

import (
	"context"
	"math"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
)

// protovalidate's defined_only means these lookups never miss.
//
//nolint:gochecknoglobals
var (
	timespans = map[yabav1.Timespan]model.Timespan{
		yabav1.Timespan_TIMESPAN_UNSPECIFIED: model.TimespanDay,
		yabav1.Timespan_TIMESPAN_DAY:         model.TimespanDay,
		yabav1.Timespan_TIMESPAN_WEEK:        model.TimespanWeek,
		yabav1.Timespan_TIMESPAN_MONTH:       model.TimespanMonth,
		yabav1.Timespan_TIMESPAN_YEAR:        model.TimespanYear,
	}
	groupBys = map[yabav1.GroupBy]model.GroupBy{
		yabav1.GroupBy_GROUP_BY_UNSPECIFIED:     model.GroupByNone,
		yabav1.GroupBy_GROUP_BY_NONE:            model.GroupByNone,
		yabav1.GroupBy_GROUP_BY_BUDGET_CATEGORY: model.GroupByBudgetCategory,
		yabav1.GroupBy_GROUP_BY_REWARD_CATEGORY: model.GroupByRewardCategory,
	}
	aggregations = map[yabav1.Aggregation]model.Aggregation{
		yabav1.Aggregation_AGGREGATION_UNSPECIFIED: model.AggregationSum,
		yabav1.Aggregation_AGGREGATION_SUM:         model.AggregationSum,
		yabav1.Aggregation_AGGREGATION_AVG:         model.AggregationAverage,
	}
)

// ExpenditureServer implements ExpenditureService for the user in the request context.
type ExpenditureServer struct {
	Pool *pgxpool.Pool
}

func (s *ExpenditureServer) ListExpenditures(
	ctx context.Context,
	req *connect.Request[yabav1.ListExpendituresRequest],
) (*connect.Response[yabav1.ListExpendituresResponse], error) {
	msg := req.Msg
	limit := 10

	if msg.Limit != nil {
		limit = int(msg.GetLimit())
	}

	var offset *int

	if msg.Offset != nil {
		offset = new(int(msg.GetOffset()))
	}

	expenditures, err := database.ListExpenditures(ctx, s.Pool, msg.Filter, msg.BudgetCategory, msg.PaymentMethodId,
		msg.Source, msg.GetSince().AsTime(), orNow(msg.GetUntil()), &limit, offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.Expenditure, 0, len(expenditures))
	for _, e := range expenditures {
		out = append(out, expenditureToProto(e))
	}

	return connect.NewResponse(&yabav1.ListExpendituresResponse{Expenditures: out}), nil
}

func (s *ExpenditureServer) AggregateExpenditures(
	ctx context.Context,
	req *connect.Request[yabav1.AggregateExpendituresRequest],
) (*connect.Response[yabav1.AggregateExpendituresResponse], error) {
	msg := req.Msg

	summaries, err := database.AggregateExpenditures(ctx, s.Pool, msg.GetSince().AsTime(), orNow(msg.GetUntil()),
		timespans[msg.GetSpan()], aggregations[msg.GetAggregation()], groupBys[msg.GetGroupBy()])
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.ExpenditureGroup, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, &yabav1.ExpenditureGroup{
			Category:  summary.Category,
			Amount:    toCents(summary.Amount),
			SpanStart: timestamppb.New(summary.StartDate),
		})
	}

	return connect.NewResponse(&yabav1.AggregateExpendituresResponse{Groups: out}), nil
}

func (s *ExpenditureServer) CreateExpenditures(
	ctx context.Context,
	req *connect.Request[yabav1.CreateExpendituresRequest],
) (*connect.Response[yabav1.CreateExpendituresResponse], error) {
	expenditures := make([]*model.Expenditure, 0, len(req.Msg.GetExpenditures()))

	for _, e := range req.Msg.GetExpenditures() {
		method := uuid.Nil
		if e.GetPaymentMethodId() != "" {
			method = uuid.MustParse(e.GetPaymentMethodId()) // validated by protovalidate
		}

		expenditures = append(expenditures, &model.Expenditure{
			Owner:          ctxutil.GetUser(ctx),
			Name:           e.GetName(),
			Amount:         float64(e.GetAmount()) / 100,
			Date:           e.GetDate().AsTime(),
			Method:         method,
			BudgetCategory: e.GetBudgetCategory(),
			RewardCategory: e.GetRewardCategory(),
			Comment:        e.GetComment(),
			Source:         e.GetSource(),
		})
	}

	if err := database.PersistExpenditures(ctx, s.Pool, expenditures); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.CreateExpendituresResponse{}), nil
}

func expenditureToProto(e *model.Expenditure) *yabav1.Expenditure {
	out := &yabav1.Expenditure{
		Id:             int64(e.ID),
		Name:           e.Name,
		Amount:         toCents(e.Amount),
		Date:           timestamppb.New(e.Date),
		BudgetCategory: e.BudgetCategory,
		RewardCategory: e.RewardCategory,
		Comment:        e.Comment,
		Source:         e.Source,
		Created:        timestamppb.New(e.CreatedTime),
	}

	if e.Method != uuid.Nil {
		out.PaymentMethodId = e.Method.String()
	}

	return out
}

func toCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func orNow(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Now()
	}

	return ts.AsTime()
}
