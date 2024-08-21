package database_test

import (
	"context"
	"testing"

	"yaba/internal/budget"
	"yaba/internal/database"
	"yaba/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBasicBudgetOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pool := helper.GetTestPool()

	owner, err := uuid.NewRandom()
	require.NoError(t, err)

	// Create and save a budget
	b := budget.NewBudget(owner, "name")
	b.SetBudgetIncome("work", 5000)
	b.SetBudgetIncome("gig", 1000)
	b.SetFixedExpense("housing", 1500)
	b.SetFixedExpense("food", 1000)
	b.SetPercentageExpense("savings", 20)
	b.SetSlackExpense("fun")
	require.NoError(t, database.PersistBudget(ctx, pool, b))

	// List budgets should show the created budget
	budgets, err := database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Len(t, budgets, 1)
	require.EqualValues(t, b, budgets[0])
	require.Len(t, budgets[0].Incomes, 2)
	require.Len(t, budgets[0].Expenses, 4)

	// Change the budget and save it
	b.RemoveExpense("savings")
	b.SetFixedExpense("dance", 200)
	b.RemoveBudgetIncome("gig")
	require.NotEqualValues(t, budgets[0], b)
	require.NoError(t, database.PersistBudget(ctx, pool, b))

	// Get the updated budget
	budgets, err = database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Len(t, budgets, 1)
	require.EqualValues(t, b, budgets[0])
	require.Len(t, budgets[0].Incomes, 1)
	require.Len(t, budgets[0].Expenses, 4)

	// Delete the budget
	require.NoError(t, database.DeleteBudget(ctx, pool, b))
	budgets, err = database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Empty(t, budgets)
}
