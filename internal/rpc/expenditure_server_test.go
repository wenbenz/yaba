package rpc_test

import (
	"testing"
	"time"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newExpenditureClient(t *testing.T, pool *pgxpool.Pool, user uuid.UUID) yabav1connect.ExpenditureServiceClient {
	t.Helper()

	srv := serve(t, pool, user)

	return yabav1connect.NewExpenditureServiceClient(srv.Client(), srv.URL+"/rpc")
}

// date returns midnight UTC of the given day in 2024, matching how DATE columns are read back.
func date(month time.Month, day int) *timestamppb.Timestamp {
	return timestamppb.New(time.Date(2024, month, day, 0, 0, 0, 0, time.UTC))
}

func TestCreateAndListExpenditures(t *testing.T) {
	t.Parallel()

	client := newExpenditureClient(t, helper.GetTestPool(), uuid.New())
	method := uuid.NewString()

	_, err := client.CreateExpenditures(t.Context(), connect.NewRequest(&yabav1.CreateExpendituresRequest{
		Expenditures: []*yabav1.NewExpenditure{
			{Date: date(3, 1), Amount: 99, Name: "Gum", PaymentMethodId: method},
			{
				Date: date(3, 3), Amount: 1234, Name: "Coffee", PaymentMethodId: method,
				BudgetCategory: "Dining", RewardCategory: "restaurant", Comment: "Oat latte", Source: "visa.csv",
			},
			{Date: date(3, 2), Amount: 500, Name: "Bus", BudgetCategory: "Transit"},
		},
	}))
	require.NoError(t, err)

	res, err := client.ListExpenditures(t.Context(), connect.NewRequest(&yabav1.ListExpendituresRequest{}))
	require.NoError(t, err)

	listed := res.Msg.GetExpenditures()
	require.Len(t, listed, 3)
	require.Positive(t, listed[0].GetId())

	coffee := &yabav1.Expenditure{
		Id: listed[0].GetId(), Name: "Coffee", Amount: 1234, Date: date(3, 3), PaymentMethodId: method,
		BudgetCategory: "Dining", RewardCategory: "restaurant", Comment: "Oat latte", Source: "visa.csv",
		Created: listed[0].GetCreated(),
	}
	require.True(t, proto.Equal(coffee, listed[0]), "listed: %v", listed[0])
	require.Empty(t, listed[1].GetPaymentMethodId(), "no payment method")

	for _, tc := range []struct {
		name string
		req  *yabav1.ListExpendituresRequest
		want []string
	}{
		{"filter matches comment", &yabav1.ListExpendituresRequest{Filter: new("LATTE")}, []string{"Coffee"}},
		{"budget category", &yabav1.ListExpendituresRequest{BudgetCategory: new("Transit")}, []string{"Bus"}},
		{"payment method", &yabav1.ListExpendituresRequest{PaymentMethodId: new(method)}, []string{"Coffee", "Gum"}},
		{"source", &yabav1.ListExpendituresRequest{Source: new("visa.csv")}, []string{"Coffee"}},
		{"date range", &yabav1.ListExpendituresRequest{Since: date(3, 2), Until: date(3, 2)}, []string{"Bus"}},
		{"limit", &yabav1.ListExpendituresRequest{Limit: new(int32(1))}, []string{"Coffee"}},
		{"offset", &yabav1.ListExpendituresRequest{Offset: new(int32(1))}, []string{"Bus", "Gum"}},
	} {
		res, err = client.ListExpenditures(t.Context(), connect.NewRequest(tc.req))
		require.NoError(t, err, tc.name)

		names := make([]string, 0, len(res.Msg.GetExpenditures()))
		for _, e := range res.Msg.GetExpenditures() {
			names = append(names, e.GetName())
		}

		require.Equal(t, tc.want, names, tc.name)
	}

	res, err = newExpenditureClient(t, helper.GetTestPool(), uuid.New()).
		ListExpenditures(t.Context(), connect.NewRequest(&yabav1.ListExpendituresRequest{}))
	require.NoError(t, err)
	require.Empty(t, res.Msg.GetExpenditures(), "other users must not see these")
}

func TestAggregateExpenditures(t *testing.T) {
	t.Parallel()

	client := newExpenditureClient(t, helper.GetTestPool(), uuid.New())

	_, err := client.CreateExpenditures(t.Context(), connect.NewRequest(&yabav1.CreateExpendituresRequest{
		Expenditures: []*yabav1.NewExpenditure{
			{Date: date(1, 10), Amount: 100, RewardCategory: "gas"},
			{Date: date(1, 10), Amount: 100, RewardCategory: "grocery"},
			{Date: date(1, 20), Amount: 102, RewardCategory: "gas"},
			{Date: date(2, 5), Amount: 300, RewardCategory: "grocery"},
		},
	}))
	require.NoError(t, err)

	month := yabav1.Timespan_TIMESPAN_MONTH

	for _, tc := range []struct {
		name string
		req  *yabav1.AggregateExpendituresRequest
		want []*yabav1.ExpenditureGroup
	}{
		{"defaults are daily totals", &yabav1.AggregateExpendituresRequest{}, []*yabav1.ExpenditureGroup{
			{Category: "Total", Amount: 200, SpanStart: date(1, 10)},
			{Category: "Total", Amount: 102, SpanStart: date(1, 20)},
			{Category: "Total", Amount: 300, SpanStart: date(2, 5)},
		}},
		{
			"monthly sum",
			&yabav1.AggregateExpendituresRequest{Span: month, Aggregation: yabav1.Aggregation_AGGREGATION_SUM},
			[]*yabav1.ExpenditureGroup{
				{Category: "Total", Amount: 302, SpanStart: date(1, 1)},
				{Category: "Total", Amount: 300, SpanStart: date(2, 1)},
			},
		},
		{
			"monthly average rounds to the nearest cent", // 302/3 = 100.67
			&yabav1.AggregateExpendituresRequest{Span: month, Aggregation: yabav1.Aggregation_AGGREGATION_AVG},
			[]*yabav1.ExpenditureGroup{
				{Category: "Total", Amount: 101, SpanStart: date(1, 1)},
				{Category: "Total", Amount: 300, SpanStart: date(2, 1)},
			},
		},
		{
			"grouped by reward category",
			&yabav1.AggregateExpendituresRequest{Span: month, GroupBy: yabav1.GroupBy_GROUP_BY_REWARD_CATEGORY},
			[]*yabav1.ExpenditureGroup{
				{Category: "gas", Amount: 202, SpanStart: date(1, 1)},
				{Category: "grocery", Amount: 100, SpanStart: date(1, 1)},
				{Category: "grocery", Amount: 300, SpanStart: date(2, 1)},
			},
		},
	} {
		res, err := client.AggregateExpenditures(t.Context(), connect.NewRequest(tc.req))
		require.NoError(t, err, tc.name)

		want := &yabav1.AggregateExpendituresResponse{Groups: tc.want}
		require.True(t, proto.Equal(want, res.Msg), "%s: %v", tc.name, res.Msg)
	}
}

func TestExpenditureService_InvalidArgument(t *testing.T) {
	t.Parallel()

	client := newExpenditureClient(t, helper.GetTestPool(), uuid.New())

	for name, req := range map[string]*yabav1.CreateExpendituresRequest{
		"empty list":   {},
		"missing date": {Expenditures: []*yabav1.NewExpenditure{{Name: "Coffee"}}},
		"malformed payment method": {
			Expenditures: []*yabav1.NewExpenditure{{Date: date(1, 1), PaymentMethodId: "not-a-uuid"}},
		},
	} {
		_, err := client.CreateExpenditures(t.Context(), connect.NewRequest(req))
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), name)
	}

	for name, req := range map[string]*yabav1.ListExpendituresRequest{
		"negative limit":           {Limit: new(int32(-1))},
		"malformed payment method": {PaymentMethodId: new("not-a-uuid")},
	} {
		_, err := client.ListExpenditures(t.Context(), connect.NewRequest(req))
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), name)
	}

	_, err := client.AggregateExpenditures(t.Context(), connect.NewRequest(
		&yabav1.AggregateExpendituresRequest{Span: yabav1.Timespan(99)}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "undefined timespan")
}

func TestExpenditureService_RequiresUser(t *testing.T) {
	t.Parallel()

	client := newExpenditureClient(t, helper.GetTestPool(), uuid.Nil)

	_, err := client.ListExpenditures(t.Context(), connect.NewRequest(&yabav1.ListExpendituresRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))

	// Auth runs before validation.
	_, err = client.CreateExpenditures(t.Context(), connect.NewRequest(&yabav1.CreateExpendituresRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestExpenditureService_DatabaseErrorIsInternal(t *testing.T) {
	t.Parallel()

	pool, err := pgxpool.NewWithConfig(t.Context(), helper.GetTestPool().Config())
	require.NoError(t, err)
	pool.Close()

	client := newExpenditureClient(t, pool, uuid.New())

	_, err = client.ListExpenditures(t.Context(), connect.NewRequest(&yabav1.ListExpendituresRequest{}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.AggregateExpenditures(t.Context(), connect.NewRequest(&yabav1.AggregateExpendituresRequest{}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.CreateExpenditures(t.Context(), connect.NewRequest(&yabav1.CreateExpendituresRequest{
		Expenditures: []*yabav1.NewExpenditure{{Date: date(1, 1)}},
	}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))
}
