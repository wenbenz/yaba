package platform_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/platform"
	"yaba/internal/test/helper"
)

func TestCSVUploadSuccess(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	path := "testdata/spend.csv"
	f, err := os.Open(path)
	require.NoError(t, err)

	user, err := uuid.NewRandom()
	require.NoError(t, err)

	ctx := ctxutil.WithUser(context.Background(), user)
	startTime := time.Now()

	require.NoError(t, platform.UploadSpendingsCSV(ctx, pool, user, f, "spend.csv"))

	endTime := time.Now()
	date, err := time.Parse(time.DateOnly, "2006-07-08")
	require.NoError(t, err)

	expenditures, err := database.ListExpenditures(ctx, pool, date, date, nil, 10)
	require.NoError(t, err)
	require.Len(t, expenditures, 3)

	require.Equal(t, "2006-07-08", expenditures[0].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 123.45, expenditures[0].Amount, .0001)
	require.Equal(t, "walmart", expenditures[0].Name)
	require.Equal(t, "debit", expenditures[0].Method)
	require.Equal(t, "groceries", expenditures[0].BudgetCategory)
	require.Equal(t, "", expenditures[0].RewardCategory)
	require.Equal(t, "", expenditures[0].Comment)
	require.Equal(t, "spend.csv", expenditures[0].Source)
	require.GreaterOrEqual(t, expenditures[0].CreatedTime, startTime)
	require.LessOrEqual(t, expenditures[0].CreatedTime, endTime)

	require.Equal(t, "2006-07-08", expenditures[2].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 99.99, expenditures[2].Amount, .0001)
	require.Equal(t, "lawn mowing", expenditures[2].Name)
	require.Equal(t, "cash", expenditures[2].Method)
	require.Equal(t, "maintenance", expenditures[2].BudgetCategory)
	require.Equal(t, "", expenditures[2].RewardCategory)
	require.Equal(t, "lawn mowing kid", expenditures[2].Comment)
	require.Equal(t, "spend.csv", expenditures[0].Source)
	require.GreaterOrEqual(t, expenditures[2].CreatedTime, startTime)
	require.LessOrEqual(t, expenditures[2].CreatedTime, endTime)
}

func TestCSVUploadBadCSV(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := context.Background()

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
		path := "testdata/" + test.filename
		f, err := os.Open(path)
		require.NoError(t, err)

		user, err := uuid.NewRandom()
		require.NoError(t, err)

		require.ErrorContains(t, platform.UploadSpendingsCSV(ctx, pool, user, f, ""), test.errorMsg)

		date, err := time.Parse(time.DateOnly, "2006-07-08")
		require.NoError(t, err)

		expenditures, err := database.ListExpenditures(ctx, pool, date, date, nil, 10)
		require.Empty(t, expenditures)
		require.NoError(t, err)
	}
}
