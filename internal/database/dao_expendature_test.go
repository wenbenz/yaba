package database_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"
	"yaba/graph/model"
	"yaba/internal/budget"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestExpenditures(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	// Create expenditures
	numExpenditures := 50
	owner := uuid.New()
	ctx := ctxutil.WithUser(context.Background(), owner)
	expenditures := make([]*budget.Expenditure, numExpenditures)

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -numExpenditures+1)

	for i := range numExpenditures {
		expenditures[i] = &budget.Expenditure{
			Owner:          owner,
			Name:           fmt.Sprintf("expenditure %d", i),
			Amount:         float64((i * 123) % 400),
			Date:           startDate.AddDate(0, 0, i),
			Method:         "cash",
			BudgetCategory: "spending",
		}
	}

	require.NoError(t, database.PersistExpenditures(ctx, pool, expenditures))

	// Fetch newly created expenditures
	fetched, err := database.ListExpenditures(ctx, pool, startDate, endDate, 100)
	require.NoError(t, err)
	require.Len(t, fetched, numExpenditures)

	// Check that they are the same
	for i, actual := range fetched {
		expected := expenditures[i]
		require.Equal(t, expected.Owner, actual.Owner)
		require.Equal(t, expected.Name, actual.Name)
		require.InDelta(t, expected.Amount, actual.Amount, .001)
		require.Equal(t, expected.Date.Format(time.DateOnly), actual.Date.Format(time.DateOnly))
		require.Equal(t, expected.BudgetCategory, actual.BudgetCategory)
		require.Equal(t, expected.Method, actual.Method)
		require.Equal(t, expected.Comment, actual.Comment)
	}

	// Fetch with smaller limit
	fetched, err = database.ListExpenditures(ctx, pool, expenditures[0].Date, endDate, 10)
	require.NoError(t, err)
	require.Equal(t, expenditures[0].Name, fetched[0].Name)
	require.Equal(t, expenditures[9].Name, fetched[9].Name)

	// Fetch with time range
	fetched, err = database.ListExpenditures(ctx, pool, fetched[4].Date, fetched[8].Date, 10)
	require.NoError(t, err)
	require.Equal(t, expenditures[4].Name, fetched[0].Name)
	require.Equal(t, expenditures[8].Name, fetched[4].Name)
}

func TestAggregateExpenditures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		dataStartDate string
		dataEndDate   string

		startDate string
		endDate   string
		span      model.Timespan
		aggregate model.Aggregation
		groupBy   model.GroupBy

		expectedAmounts func(expenditures []*budget.Expenditure) []float64
	}{
		{
			name:          "timespan week",
			dataStartDate: "2024-08-05",
			dataEndDate:   "2024-08-18",

			startDate: "2024-08-05",
			endDate:   "2024-08-18",
			span:      model.TimespanWeek,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByNone,

			expectedAmounts: twoWeeksInAugust(),
		}, {
			name:          "date range subset of data",
			dataStartDate: "2022-08-01",
			dataEndDate:   "2024-08-30",

			startDate: "2024-08-05",
			endDate:   "2024-08-18",
			span:      model.TimespanWeek,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByNone,

			expectedAmounts: twoWeeksInAugust(),
		}, {
			name:          "date range superset of data",
			dataStartDate: "2024-08-05",
			dataEndDate:   "2024-08-18",

			startDate: "2024-08-01",
			endDate:   "2024-08-30",
			span:      model.TimespanWeek,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByNone,

			expectedAmounts: twoWeeksInAugust(),
		}, {
			name:          "day span",
			dataStartDate: "2024-08-01",
			dataEndDate:   "2024-08-10",

			startDate: "2024-08-01",
			endDate:   "2024-08-30",
			span:      model.TimespanDay,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByNone,

			expectedAmounts: func(expenditures []*budget.Expenditure) []float64 {
				out := make([]float64, 10)
				for _, e := range expenditures {
					out[e.Date.Day()-1] += e.Amount
				}

				return out
			},
		}, {
			name:          "month span",
			dataStartDate: "2024-01-01",
			dataEndDate:   "2024-03-10",

			startDate: "2024-01-01",
			endDate:   "2024-02-28",
			span:      model.TimespanMonth,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByNone,

			expectedAmounts: func(expenditures []*budget.Expenditure) []float64 {
				out := make([]float64, 2)

				for _, e := range expenditures {
					out[e.Date.Month()-1] += e.Amount
				}

				return out
			},
		}, {
			name:          "aggregate average",
			dataStartDate: "2024-01-01",
			dataEndDate:   "2024-12-31",

			startDate: "2024-01-01",
			endDate:   "2024-12-31",
			span:      model.TimespanYear,
			aggregate: model.AggregationAvg,
			groupBy:   model.GroupByNone,

			expectedAmounts: func(expenditures []*budget.Expenditure) []float64 {
				sum := 0.

				for _, e := range expenditures {
					sum += e.Amount
				}

				return []float64{sum / float64(len(expenditures))}
			},
		}, {
			name:          "group by budget category",
			dataStartDate: "2024-01-01",
			dataEndDate:   "2024-02-28",

			startDate: "2024-01-01",
			endDate:   "2024-02-28",
			span:      model.TimespanMonth,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByBudgetCategory,

			expectedAmounts: janFebGroupBy(model.GroupByBudgetCategory),
		}, {
			name:          "group by reward category",
			dataStartDate: "2024-01-01",
			dataEndDate:   "2024-02-28",

			startDate: "2024-01-01",
			endDate:   "2024-02-28",
			span:      model.TimespanMonth,
			aggregate: model.AggregationSum,
			groupBy:   model.GroupByRewardCategory,

			expectedAmounts: janFebGroupBy(model.GroupByRewardCategory),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup data
			user := uuid.New()
			ctx := ctxutil.WithUser(context.Background(), user)
			pool := helper.GetTestPool()

			startDate, _ := time.Parse(time.DateOnly, tc.dataStartDate)
			endDate, _ := time.Parse(time.DateOnly, tc.dataEndDate)

			err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
			require.NoError(t, err)

			// Make the call
			startDate, _ = time.Parse(time.DateOnly, tc.startDate)
			endDate, _ = time.Parse(time.DateOnly, tc.endDate)
			aggregate, err := database.AggregateExpenditures(ctx, pool, startDate, endDate, tc.span, tc.aggregate, tc.groupBy)
			require.NoError(t, err)

			// Calculate the expected and compare the amounts
			expenditures, err := database.ListExpenditures(ctx, pool, startDate, endDate, 300)
			require.NoError(t, err)

			expected := tc.expectedAmounts(expenditures)
			require.Len(t, aggregate, len(expected))

			for i := range aggregate {
				require.InDelta(t, expected[i], aggregate[i].Amount, .001)
			}
		})
	}
}

func janFebGroupBy(groupBy model.GroupBy) func(expenditures []*budget.Expenditure) []float64 {
	return func(expenditures []*budget.Expenditure) []float64 {
		buckets := []map[string]float64{{}, {}}
		exists := map[string]bool{}

		var categories []string

		for _, e := range expenditures {
			var category string
			//nolint:exhaustive
			switch groupBy {
			case model.GroupByBudgetCategory:
				category = e.BudgetCategory
			case model.GroupByRewardCategory:
				category = e.RewardCategory
			}

			buckets[e.Date.Month()-1][category] += e.Amount

			if _, ok := exists[category]; !ok {
				categories = append(categories, category)
				exists[category] = true
			}
		}

		slices.Sort(categories)

		var out []float64

		for _, bucket := range buckets {
			for _, cat := range categories {
				if sum, ok := bucket[cat]; ok {
					out = append(out, sum)
				}
			}
		}

		return out
	}
}

func twoWeeksInAugust() func(expenditures []*budget.Expenditure) []float64 {
	return func(expenditures []*budget.Expenditure) []float64 {
		monday, _ := time.Parse(time.DateOnly, "2024-08-12")
		out := make([]float64, 2)

		for _, e := range expenditures {
			if e.Date.Before(monday) {
				out[0] += e.Amount
			} else {
				out[1] += e.Amount
			}
		}

		return out
	}
}
