package database_test

import (
	"github.com/brianvoe/gofakeit"
	"slices"
	"sort"
	"testing"
	"time"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListAllExpenditures(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	owner := uuid.New()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	generated := helper.MockExpenditures(10, owner, startDate, endDate)

	// Sort expenditures and make sure dates are unique
	sort.Slice(generated, func(i, j int) bool {
		return generated[i].Date.After(generated[j].Date)
	})

	expenditures := make([]*model.Expenditure, 0, len(generated))
	lastDate := time.Time{}

	for _, exp := range generated {
		if exp.Date.Equal(lastDate) {
			continue
		}

		expenditures = append(expenditures, exp)
		lastDate = exp.Date
	}

	require.NotEmptyf(t, expenditures, "expenditures should not be empty")

	require.NoError(t, database.PersistExpenditures(t.Context(), pool, expenditures))

	fetched, err := database.ListExpenditures(
		ctxutil.WithUser(t.Context(), owner), pool, startDate, endDate, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, fetched, len(expenditures))

	for i, exp := range fetched {
		require.Equal(t, exp.Owner, owner)
		require.Equal(t, exp.Name, expenditures[i].Name)
		require.InDelta(t, exp.Amount, expenditures[i].Amount, .001)
		require.Equal(t, expenditures[i].Date.Format(time.DateOnly), exp.Date.Format(time.DateOnly))
		require.Equal(t, exp.BudgetCategory, expenditures[i].BudgetCategory)
		require.Equal(t, exp.Method, expenditures[i].Method)
		require.Equal(t, exp.Comment, expenditures[i].Comment)
	}
}

//nolint:gocognit,cyclop
func TestListExpenditures(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	// Create expenditures
	numExpenditures := 500
	owner := uuid.New()
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -30)
	generated := helper.MockExpenditures(numExpenditures, owner, startDate, endDate)

	// Randomly make 10 of them have no categories
	for range 10 {
		for {
			j := gofakeit.Number(0, numExpenditures)
			if generated[j].BudgetCategory != "" {
				generated[j].BudgetCategory = ""

				break
			}
		}
	}

	// Persist expenditures
	require.NoError(t, database.PersistExpenditures(t.Context(), pool, generated))

	// Refetch expenditures so IDs and order is correct for other tests.
	// TestListAllExpenditures verifies the integrity of using this.
	expenditures, err := database.ListExpenditures(
		ctxutil.WithUser(t.Context(), owner), pool, startDate, endDate, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, expenditures, numExpenditures, "Num expenditures doesn't match")

	testCases := []struct {
		name       string
		queryStart time.Time
		queryEnd   time.Time
		category   *string
		source     *string
		cursor     *int
		limit      *int
		expected   func() []*model.Expenditure
	}{
		{
			name:       "fetch_with_limit",
			queryStart: startDate,
			queryEnd:   endDate,
			category:   nil,
			source:     nil,
			cursor:     nil,
			limit:      pointer(10),
			expected: func() []*model.Expenditure {
				return expenditures[:10]
			},
		},
		{
			name:       "fetch_with_cursor",
			queryStart: startDate,
			queryEnd:   endDate,
			category:   nil,
			source:     nil,
			cursor:     pointer(10),
			limit:      pointer(10),
			expected: func() []*model.Expenditure {
				return expenditures[10:20]
			},
		},
		{
			name:       "fetch_by_date_range",
			queryStart: startDate.AddDate(0, 0, 5),
			queryEnd:   startDate.AddDate(0, 0, 10),
			category:   nil,
			source:     nil,
			cursor:     nil,
			limit:      nil,
			expected: func() []*model.Expenditure {
				var result []*model.Expenditure
				for _, e := range expenditures {
					if !e.Date.Before(startDate.AddDate(0, 0, 5)) &&
						!e.Date.After(startDate.AddDate(0, 0, 10)) {
						result = append(result, e)
					}
				}
				require.NotEmpty(t, result, "expenditures should not be empty")

				return result
			},
		},
		{
			name:       "fetch_by_category",
			queryStart: startDate,
			queryEnd:   endDate,
			category:   pointer("Electric"),
			source:     nil,
			cursor:     nil,
			limit:      nil,
			expected: func() []*model.Expenditure {
				var result []*model.Expenditure
				for _, e := range expenditures {
					if e.BudgetCategory == "Electric" {
						result = append(result, e)
					}
				}
				require.NotEmpty(t, result, "expenditures should not be empty")

				return result
			},
		},
		{
			name:       "fetch_by_empty_category",
			queryStart: startDate,
			queryEnd:   endDate,
			category:   pointer(""),
			source:     nil,
			cursor:     nil,
			limit:      nil,
			expected: func() []*model.Expenditure {
				var result []*model.Expenditure
				for _, e := range expenditures {
					if e.BudgetCategory == "" {
						result = append(result, e)
					}
				}

				require.Len(t, result, 10)

				return result
			},
		}, {
			name:       "fetch_by_source",
			queryStart: startDate,
			queryEnd:   endDate,
			category:   nil,
			source:     pointer("Nissan.csv"),
			cursor:     nil,
			limit:      pointer(100),
			expected: func() []*model.Expenditure {
				var result []*model.Expenditure
				for _, e := range expenditures {
					if e.Source == "Nissan.csv" {
						result = append(result, e)
					}
				}
				require.NotEmpty(t, result, "expenditures should not be empty")

				return result
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := ctxutil.WithUser(t.Context(), owner)

			// Run test case
			fetched, err := database.ListExpenditures(
				ctx, pool, tc.queryStart, tc.queryEnd, tc.source, tc.category, tc.limit, tc.cursor)
			require.NoError(t, err)

			expected := tc.expected()
			require.Len(t, fetched, len(expected))

			// Verify results
			for i, actual := range fetched {
				exp := expected[i]
				require.Equal(t, exp.Owner, actual.Owner)
				require.Equal(t, exp.Name, actual.Name)
				require.InDelta(t, exp.Amount, actual.Amount, .001)
				require.Equal(t, exp.Date.Format(time.DateOnly), actual.Date.Format(time.DateOnly))
				require.Equal(t, exp.BudgetCategory, actual.BudgetCategory)
				require.Equal(t, exp.Method, actual.Method)
				require.Equal(t, exp.Comment, actual.Comment)
			}
		})
	}
}

func pointer[T any](v T) *T {
	return &v
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

		expectedAmounts func(expenditures []*model.Expenditure) []float64
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

			expectedAmounts: func(expenditures []*model.Expenditure) []float64 {
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

			expectedAmounts: func(expenditures []*model.Expenditure) []float64 {
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
			aggregate: model.AggregationAverage,
			groupBy:   model.GroupByNone,

			expectedAmounts: func(expenditures []*model.Expenditure) []float64 {
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
			ctx := ctxutil.WithUser(t.Context(), user)
			pool := helper.GetTestPool()

			startDate, _ := time.ParseInLocation(time.DateOnly, tc.dataStartDate, time.UTC)
			endDate, _ := time.ParseInLocation(time.DateOnly, tc.dataEndDate, time.UTC)

			mocked := helper.MockExpenditures(300, user, startDate, endDate)

			// Create a basic budget with all the generated categories
			budget := model.NewBudget(user, "test budget")

			categories := make(map[string]bool)
			for i, e := range mocked {
				if e.BudgetCategory != "" && !categories[e.BudgetCategory] {
					budget.SetBasicExpense(e.BudgetCategory, float64(i*100))

					categories[e.BudgetCategory] = true
				}
			}

			require.NoError(t, database.PersistBudget(ctx, pool, budget))

			// Persist expenditures after budget so we can group by budget category
			err := database.PersistExpenditures(ctx, pool, mocked)
			require.NoError(t, err)

			// Make the call
			startDate, _ = time.ParseInLocation(time.DateOnly, tc.startDate, time.UTC)
			endDate, _ = time.ParseInLocation(time.DateOnly, tc.endDate, time.UTC)
			aggregate, err := database.AggregateExpenditures(ctx, pool, startDate, endDate, tc.span, tc.aggregate, tc.groupBy)
			require.NoError(t, err)

			// Calculate the expected and compare the amounts
			expenditures, err := database.ListExpenditures(ctx, pool, startDate, endDate, nil, nil, pointer(300), nil)
			require.NoError(t, err)

			expected := tc.expectedAmounts(expenditures)
			require.Len(t, aggregate, len(expected))

			for i := range aggregate {
				require.InDelta(t, expected[i], aggregate[i].Amount, .001)
			}
		})
	}
}

func janFebGroupBy(groupBy model.GroupBy) func(expenditures []*model.Expenditure) []float64 {
	return func(expenditures []*model.Expenditure) []float64 {
		buckets := []map[string]float64{{}, {}}
		exists := map[string]bool{}

		var categories []string

		for _, e := range expenditures {
			var category string
			//nolint:exhaustive
			switch groupBy {
			case model.GroupByBudgetCategory:
				category = e.ExpenseID.String()
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

func twoWeeksInAugust() func(expenditures []*model.Expenditure) []float64 {
	return func(expenditures []*model.Expenditure) []float64 {
		monday, _ := time.ParseInLocation(time.DateOnly, "2024-08-12", time.UTC)
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
