package importer_test

import (
	"encoding/csv"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
	importer "yaba/internal/import"
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
	}

	for _, test := range testCases {
		path := "testdata/" + test.filename
		rows := test.expectedRows
		f, err := os.Open(path)
		require.NoError(t, err)

		owner, err := uuid.NewRandom()
		require.NoError(t, err)

		csvReader := csv.NewReader(f)
		expenditures, err := importer.ImportExpendituresFromCSVReader(owner, csvReader)
		require.NoError(t, err)
		require.Len(t, expenditures, rows)
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
			filename: "extra_column.csv",
			errorMsg: "unrecognized column 'pineapple'",
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
	}

	for _, test := range testCases {
		path := "testdata/" + test.filename
		f, err := os.Open(path)
		require.NoError(t, err)

		owner, err := uuid.NewRandom()
		require.NoError(t, err)

		csvReader := csv.NewReader(f)
		_, err = importer.ImportExpendituresFromCSVReader(owner, csvReader)
		require.ErrorContains(t, err, test.errorMsg, "failing test: "+test.filename)
	}
}

func TestCsvExpenditureReader(t *testing.T) {
	t.Parallel()

	headers := []string{"date", "amount", "name", "method", "budget_category", "reward_category", "comment"}
	owner, _ := uuid.NewRandom()
	reader, err := importer.NewCSVExpenditureReader(owner, headers)
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
