package importer_test

import (
	"encoding/csv"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	importer "yaba/internal/import"
	"yaba/internal/test/helper"
)

func TestValidCSVs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		filename     string
		expectedRows int
	}{
		{
			filename:     "all_cols.csv",
			expectedRows: 1,
		},
		{
			filename:     "acceptable_dollars.csv",
			expectedRows: 6,
		},
		{
			filename:     "extra_column.csv",
			expectedRows: 1,
		},
		{
			filename:     "amex.csv",
			expectedRows: 1,
		},
		{
			filename:     "columns_only.csv",
			expectedRows: 0,
		},
		{
			filename:     "empty_row.csv",
			expectedRows: 1,
		},
	}

	for _, test := range testCases {
		t.Run("CSV:"+test.filename, func(t *testing.T) {
			t.Parallel()

			path := "testdata/" + test.filename
			rows := test.expectedRows
			f, err := os.Open(path)
			require.NoError(t, err)

			owner, err := uuid.NewRandom()
			require.NoError(t, err)

			csvReader := csv.NewReader(f)
			expenditures, err := importer.ImportExpendituresFromCSVReader(owner, "testSource", csvReader)
			require.NoError(t, err)
			require.Len(t, expenditures, rows)
		})
	}
}

func TestInvalidCSVs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		filename string
		errorMsg string
	}{
		{
			filename: "empty.csv",
			errorMsg: "received error reading headers: EOF",
		},
		{
			filename: "missing_amount.csv",
			errorMsg: "missing required column 'amount'",
		},
		{
			filename: "missing_date.csv",
			errorMsg: "missing required column 'date'",
		},
		{
			filename: "unparsable_amount.csv",
			errorMsg: "failed to parse dollars from '123gg.45",
		},
		{
			filename: "unparsable_date.csv",
			errorMsg: "date must have format YYYY-MM-DD",
		},
		{
			filename: "incomplete_row.csv",
			errorMsg: "unexpected error reading csv: record on line 2: wrong number of fields",
		},
	}

	for _, test := range testCases {
		t.Run("CSV:"+test.filename, func(t *testing.T) {
			t.Parallel()

			path := "testdata/" + test.filename
			f, err := os.Open(path)
			require.NoError(t, err)

			owner, err := uuid.NewRandom()
			require.NoError(t, err)

			csvReader := csv.NewReader(f)
			_, err = importer.ImportExpendituresFromCSVReader(owner, "testSource", csvReader)
			require.ErrorContains(t, err, test.errorMsg, "failing test: "+test.filename)
		})
	}
}

func TestCsvExpenditureReader(t *testing.T) {
	t.Parallel()

	headers := []string{"date", "amount", "name", "method", "budget_category", "reward_category", "comment"}
	owner, _ := uuid.NewRandom()
	reader, err := importer.NewCSVExpenditureReader(owner, "testSource", headers)
	require.NoError(t, err)

	date, amount, name, method, budgetCategory, rewardCategory, comment :=
		"2029-08-09", "12,345.67", "nuclear bunkers inc.", "gold bars", "shelter", "", ""

	row := []string{date, amount, name, method, budgetCategory, rewardCategory, comment}
	expenditure, err := reader.ReadRow(row)
	require.NoError(t, err)

	require.Equal(t, date, expenditure.Date.Format(time.DateOnly))
	require.InDelta(t, 12345.67, expenditure.Amount, .001)
	require.Equal(t, name, expenditure.Name)
	require.Equal(t, method, expenditure.Method)
	require.Equal(t, budgetCategory, expenditure.BudgetCategory)
	require.Equal(t, rewardCategory, expenditure.RewardCategory)
	require.Equal(t, comment, expenditure.Comment)
}

func TestCSVUploadSuccess(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	path := "testdata/spend.csv"
	f, err := os.Open(path)
	require.NoError(t, err)

	user, err := uuid.NewRandom()
	require.NoError(t, err)

	ctx := ctxutil.WithUser(t.Context(), user)
	startTime := time.Now()

	require.NoError(t, importer.UploadSpendingsCSV(ctx, pool, user, f, "spend.csv"))

	endTime := time.Now()
	date, err := time.Parse(time.DateOnly, "2006-07-08")
	require.NoError(t, err)

	expenditures, err := database.ListExpenditures(ctx, pool, date, date, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, expenditures, 3)

	require.Equal(t, "2006-07-08", expenditures[2].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 123.45, expenditures[2].Amount, .0001)
	require.Equal(t, "walmart", expenditures[2].Name)
	require.Equal(t, "debit", expenditures[2].Method)
	require.Equal(t, "groceries", expenditures[2].BudgetCategory)
	require.Equal(t, "", expenditures[2].RewardCategory)
	require.Equal(t, "", expenditures[2].Comment)
	require.Equal(t, "spend.csv", expenditures[2].Source)
	require.GreaterOrEqual(t, expenditures[2].CreatedTime, startTime)
	require.LessOrEqual(t, expenditures[2].CreatedTime, endTime)

	require.Equal(t, "2006-07-08", expenditures[0].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 99.99, expenditures[0].Amount, .0001)
	require.Equal(t, "lawn mowing", expenditures[0].Name)
	require.Equal(t, "cash", expenditures[0].Method)
	require.Equal(t, "maintenance", expenditures[0].BudgetCategory)
	require.Equal(t, "", expenditures[0].RewardCategory)
	require.Equal(t, "lawn mowing kid", expenditures[0].Comment)
	require.Equal(t, "spend.csv", expenditures[0].Source)
	require.GreaterOrEqual(t, expenditures[0].CreatedTime, startTime)
	require.LessOrEqual(t, expenditures[0].CreatedTime, endTime)
}

func TestCSVUploadBadCSV(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := t.Context()

	testCases := []struct {
		filename string
		errorMsg string
	}{
		{
			filename: "invalid.csv",
			errorMsg: "failed to import: failed to parse dollars from '12.b'",
		},
		{
			filename: "invalid_category.csv",
			errorMsg: "invalid input value for enum reward_category",
		},
	}
	for _, test := range testCases {
		t.Run("CSV:"+test.filename, func(t *testing.T) {
			t.Parallel()

			path := "testdata/" + test.filename
			f, err := os.Open(path)
			require.NoError(t, err)

			user, err := uuid.NewRandom()
			require.NoError(t, err)

			require.ErrorContains(t, importer.UploadSpendingsCSV(ctx, pool, user, f, ""), test.errorMsg)

			date, err := time.Parse(time.DateOnly, "2006-07-08")
			require.NoError(t, err)

			expenditures, err := database.ListExpenditures(ctx, pool, date, date, nil, nil, nil, nil)
			require.Empty(t, expenditures)
			require.NoError(t, err)
		})
	}
}
