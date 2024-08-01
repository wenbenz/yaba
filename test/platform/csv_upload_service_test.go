package platform_test

import (
	"context"
	"os"
	"testing"
	"time"
	"yaba/internal/database"
	"yaba/internal/platform"
	"yaba/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCSVUploadSuccess(t *testing.T) {
	t.Parallel()

	pool, closeContainer := helper.SetupTestContainerAndInitPool()
	defer closeContainer()

	path := "testdata/spend.csv"
	f, err := os.Open(path)
	require.NoError(t, err)

	user, err := uuid.NewRandom()
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "user", user.String())

	require.NoError(t, platform.UploadSpendingsCSV(ctx, pool, f))

	date, err := time.Parse(time.DateOnly, "2006-07-08")
	require.NoError(t, err)

	expenditures, err := database.ListExpenditures(ctx, pool, user, date, date, 10)
	require.NoError(t, err)
	require.Len(t, expenditures, 3)

	require.Equal(t, "2006-07-08", expenditures[0].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 123.45, expenditures[0].Amount, .0001)
	require.Equal(t, "walmart", expenditures[0].Name)
	require.Equal(t, "debit", expenditures[0].Method)
	require.Equal(t, "groceries", expenditures[0].BudgetCategory)
	require.Equal(t, "", expenditures[0].RewardCategory)
	require.Equal(t, "", expenditures[0].Comment)

	require.Equal(t, "2006-07-08", expenditures[1].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 12., expenditures[1].Amount, .0001)
	require.Equal(t, "target", expenditures[1].Name)
	require.Equal(t, "credit", expenditures[1].Method)
	require.Equal(t, "groceries", expenditures[1].BudgetCategory)
	require.Equal(t, "retail", expenditures[1].RewardCategory)
	require.Equal(t, "", expenditures[1].Comment)

	require.Equal(t, "2006-07-08", expenditures[2].Date.UTC().Format(time.DateOnly))
	require.InDelta(t, 99.99, expenditures[2].Amount, .0001)
	require.Equal(t, "lawn mowing", expenditures[2].Name)
	require.Equal(t, "cash", expenditures[2].Method)
	require.Equal(t, "maintenance", expenditures[2].BudgetCategory)
	require.Equal(t, "", expenditures[2].RewardCategory)
	require.Equal(t, "lawn mowing kid", expenditures[2].Comment)
}

func TestCSVUploadBadCSV(t *testing.T) {
	t.Parallel()

	pool, closeContainer := helper.SetupTestContainerAndInitPool()
	defer closeContainer()

	path := "testdata/invalid.csv"
	f, err := os.Open(path)
	require.NoError(t, err)

	user, err := uuid.NewRandom()
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "user", user.String())

	require.ErrorContains(t, platform.UploadSpendingsCSV(ctx, pool, f), "failed to import: failed to parse dollars from '12.b'")

	date, err := time.Parse(time.DateOnly, "2006-07-08")
	require.NoError(t, err)

	expenditures, err := database.ListExpenditures(ctx, pool, user, date, date, 10)
	require.Len(t, expenditures, 0)
	require.NoError(t, err)
}

func TestCSVUploadNoUser(t *testing.T) {
	t.Parallel()

	pool, closeContainer := helper.SetupTestContainerAndInitPool()
	defer closeContainer()

	path := "testdata/spend.csv"
	f, err := os.Open(path)
	require.NoError(t, err)

	require.ErrorContains(t, platform.UploadSpendingsCSV(context.Background(), pool, f), "no user in context")
}
