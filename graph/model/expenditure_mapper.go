package model

import (
	"fmt"
	"strconv"
	"time"
	"yaba/internal/model"
)

func ExpendituresToExpenitureResponse(expenditures []*model.Expenditure) []*ExpenditureResponse {
	ret := make([]*ExpenditureResponse, len(expenditures))

	for i, obj := range expenditures {
		id := strconv.Itoa(obj.ID)
		owner := obj.Owner.String()
		amount := fmt.Sprintf("%.2f", obj.Amount)
		date := obj.Date.Format(time.DateOnly)
		cat := obj.RewardCategory
		created := obj.CreatedTime.Format(time.DateOnly)

		ret[i] = &ExpenditureResponse{
			ID:             &id,
			Owner:          &owner,
			Name:           &obj.Name,
			Amount:         &amount,
			Date:           &date,
			Method:         &obj.Method,
			BudgetCategory: &obj.BudgetCategory,
			RewardCategory: &cat,
			Comment:        &obj.Comment,
			Created:        &created,
			Source:         &obj.Source,
		}
	}

	return ret
}

func ExpenditureSummariesToAggregateExpenditures(expenditures []*model.ExpenditureSummary, timespan Timespan,
) []*AggregatedExpendituresResponse {
	ret := make([]*AggregatedExpendituresResponse, len(expenditures))

	for i, obj := range expenditures {
		start := obj.StartDate.Format(time.DateOnly)
		ret[i] = &AggregatedExpendituresResponse{
			GroupByCategory: &obj.Category,
			Amount:          &obj.Amount,
			SpanStart:       &start,
			Span:            &timespan,
		}
	}

	return ret
}
func ConvertAggregation(agg Aggregation) model.Aggregation {
	switch agg {
	case AggregationSum:
		return model.AggregationSum
	case AggregationAvg:
		return model.AggregationAverage
	default:
		return model.AggregationSum
	}
}

func ConvertTimespan(span Timespan) model.Timespan {
	switch span {
	case TimespanDay:
		return model.TimespanDay
	case TimespanWeek:
		return model.TimespanWeek
	case TimespanMonth:
		return model.TimespanMonth
	case TimespanYear:
		return model.TimespanYear
	default:
		return model.TimespanDay
	}
}

func ConvertGroupBy(groupBy GroupBy) model.GroupBy {
	switch groupBy {
	case GroupByNone:
		return model.GroupByNone
	case GroupByBudgetCategory:
		return model.GroupByBudgetCategory
	case GroupByRewardCategory:
		return model.GroupByRewardCategory
	default:
		return model.GroupByNone
	}
}
