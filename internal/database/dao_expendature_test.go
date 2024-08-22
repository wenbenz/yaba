package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"yaba/internal/database"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"yaba/internal/budget"
)

func TestExpenditures(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	ctx := context.Background()

	// Create expenditures
	numExpenditures := 50
	owner, _ := uuid.NewRandom()
	expenditures := make([]*budget.Expenditure, numExpenditures)

	for i := range numExpenditures {
		expenditures[i] = &budget.Expenditure{
			Owner:          owner,
			Name:           fmt.Sprintf("expenditure %d", i),
			Amount:         float64((i * 123) % 400),
			Date:           time.Now().Add(time.Duration(i-numExpenditures) * time.Hour),
			Method:         "cash",
			BudgetCategory: "spending",
		}
	}

	require.NoError(t, database.PersistExpenditures(ctx, pool, expenditures))

	// Fetch newly created expenditures
	fetched, err := database.ListExpenditures(ctx, pool, owner,
		time.Now().Add(time.Duration(-52)*time.Hour), time.Now(), 100)
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
		require.Equal(t, expected.RewardCategory, actual.RewardCategory)
		require.Equal(t, expected.Method, actual.Method)
		require.Equal(t, expected.Comment, actual.Comment)
	}

	// Fetch with smaller limit
	fetched, err = database.ListExpenditures(ctx, pool, owner, expenditures[0].Date, time.Now(), 10)
	require.NoError(t, err)
	require.Equal(t, expenditures[0].Name, fetched[0].Name)
	require.Equal(t, expenditures[9].Name, fetched[9].Name)

	// Fetch with time range
	fetched, err = database.ListExpenditures(ctx, pool, owner, fetched[4].Date, fetched[8].Date, 10)
	require.NoError(t, err)
	require.Equal(t, expenditures[4].Name, fetched[0].Name)
	require.Equal(t, expenditures[8].Name, fetched[4].Name)
}
