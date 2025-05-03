package database_test

import (
	"testing"
	"time"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBasicBudgetOperations(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	owner := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), owner)

	// Create and save a budget
	b := model.NewBudget(owner, "name")
	b.SetBudgetIncome("work", 5000)
	b.SetBudgetIncome("gig", 1000)
	b.SetFixedExpense("housing", 1500)
	b.SetFixedExpense("food", 1000)
	b.SetBasicExpense("savings", 20)
	b.SetSlackExpense("fun")
	require.NoError(t, database.PersistBudget(ctx, pool, b))

	// List budgets should show the created budget
	budgets, err := database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Len(t, budgets, 1)

	for i, expense := range budgets[0].Expenses {
		require.NotEqual(t, uuid.Nil, expense.ID)
		b.Expenses[i].ID = expense.ID
	}

	require.Equal(t, b, budgets[0])
	require.Len(t, budgets[0].Incomes, 2)
	require.Len(t, budgets[0].Expenses, 4)

	// Get specific budget
	fetched, err := database.GetBudget(ctx, pool, owner, b.ID)
	require.NoError(t, err)
	require.Equal(t, b, fetched)

	// Change the budget and save it
	b.RemoveExpense("savings")
	b.SetFixedExpense("dance", 200)
	b.RemoveBudgetIncome("gig")
	require.NotEqual(t, budgets[0], b)
	require.NoError(t, database.PersistBudget(ctx, pool, b))

	// Get the updated budget
	budgets, err = database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Len(t, budgets, 1)
	b.Expenses[3].ID = budgets[0].Expenses[3].ID
	require.Equal(t, b, budgets[0])
	require.Len(t, budgets[0].Incomes, 1)
	require.Len(t, budgets[0].Expenses, 4)

	// Delete the budget
	require.NoError(t, database.DeleteBudget(ctx, pool, b))
	budgets, err = database.GetBudgets(ctx, pool, owner, 10)
	require.NoError(t, err)
	require.Empty(t, budgets)
}

func TestGetNonExistingBudget(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	pool := helper.GetTestPool()

	// Get specific budget
	_, err := database.GetBudget(ctx, pool, uuid.New(), uuid.New())
	require.ErrorContains(t, err, "no such element")
}

func TestPersistUnauthorizedBudget(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	owner := uuid.New()
	unauthorized := uuid.New()

	// Create initial budget with original owner
	b := model.NewBudget(owner, "original")
	b.SetBudgetIncome("work", 1000)

	ctx := ctxutil.WithUser(t.Context(), owner)
	require.NoError(t, database.PersistBudget(ctx, pool, b))

	// Attempt to update budget with unauthorized user
	b.Name = "modified"
	ctx = ctxutil.WithUser(t.Context(), unauthorized)
	err := database.PersistBudget(ctx, pool, b)
	require.Error(t, err) // Should succeed but not modify

	// Verify budget remains unchanged
	budget, err := database.GetBudget(ctxutil.WithUser(t.Context(), owner), pool, owner, b.ID)
	require.NoError(t, err)
	require.Equal(t, "original", budget.Name)
	require.Equal(t, owner, budget.Owner)
}

func TestPersistBudgetClassifiesExpenditures(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	owner := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), owner)

	// Create initial expenditures with categories
	existingExpenseID := uuid.New()
	expenditures := []*model.Expenditure{
		{
			Owner:          owner,
			Name:           "Walmart",
			Amount:         50.00,
			Date:           time.Now(),
			BudgetCategory: "groceries",
			ExpenseID:      existingExpenseID, // Already classified
		},
		{
			Owner:          owner,
			Name:           "Target",
			Amount:         30.00,
			Date:           time.Now(),
			BudgetCategory: "groceries", // Unclassified
		},
		{
			Owner:          owner,
			Name:           "Netflix",
			Amount:         15.00,
			Date:           time.Now(),
			BudgetCategory: "entertainment", // Unclassified
		},
	}
	require.NoError(t, database.PersistExpenditures(ctx, pool, expenditures))

	// Create and persist a new budget
	budget := model.NewBudget(owner, "test budget")
	budget.SetBasicExpense("groceries", 500)
	budget.SetBasicExpense("entertainment", 100)
	require.NoError(t, database.PersistBudget(ctx, pool, budget))

	// Verify expenditures were updated
	fetched, err := database.ListExpenditures(ctx, pool,
		nil, nil, nil, nil, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1), nil, nil)
	require.NoError(t, err)
	require.Len(t, fetched, 3)

	// Find the new expense IDs from the budget
	groceryExpense := findExpenseByCategory(budget.Expenses, "groceries")
	require.NotNil(t, groceryExpense)

	entertainmentExpense := findExpenseByCategory(budget.Expenses, "entertainment")
	require.NotNil(t, entertainmentExpense)

	// Verify each expenditure
	for _, e := range fetched {
		switch e.BudgetCategory {
		case "groceries":
			if e.Name == "Walmart" {
				require.Equal(
					t,
					existingExpenseID,
					e.ExpenseID,
				) // Should keep existing classification
			} else {
				require.Equal(t, groceryExpense.ID, e.ExpenseID) // Should be newly classified
			}
		case "entertainment":
			require.Equal(t, entertainmentExpense.ID, e.ExpenseID) // Should be newly classified
		}
	}
}

func TestPersistBudgetClassifiesExpendituresWithExistingBudget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		initialExpenditures []*model.Expenditure
		existingExpenses    []string
		newExpenses         []string
		expectedMatches     map[string]int // category -> count of classified expenditures
	}{
		{
			name: "new category added",
			initialExpenditures: []*model.Expenditure{
				{Name: "Walmart", Amount: 50.00, BudgetCategory: "groceries"},
				{Name: "Netflix", Amount: 15.00, BudgetCategory: "entertainment"},
			},
			existingExpenses: []string{"groceries"},
			newExpenses:      []string{"entertainment"},
			expectedMatches: map[string]int{
				"groceries":     1,
				"entertainment": 1,
			},
		},
		{
			name: "multiple unclassified for same category",
			initialExpenditures: []*model.Expenditure{
				{Name: "Walmart", Amount: 50.00, BudgetCategory: "groceries"},
				{Name: "Target", Amount: 30.00, BudgetCategory: "groceries"},
				{Name: "Corner Store", Amount: 10.00, BudgetCategory: "groceries"},
			},
			existingExpenses: []string{"rent"},
			newExpenses:      []string{"groceries"},
			expectedMatches: map[string]int{
				"groceries": 3,
			},
		},
		{
			name: "mixed classified and unclassified",
			initialExpenditures: []*model.Expenditure{
				{
					Name:           "Walmart",
					Amount:         50.00,
					BudgetCategory: "groceries",
				}, // Will be classified
				{
					Name:           "Netflix",
					Amount:         15.00,
					BudgetCategory: "entertainment",
				}, // Will be classified
				{Name: "Rent", Amount: 12.00, BudgetCategory: "rent"},
			},
			existingExpenses: []string{"rent"},
			newExpenses:      []string{"groceries", "entertainment"},
			expectedMatches: map[string]int{
				"groceries":     1,
				"entertainment": 1,
				"rent":          1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pool := helper.GetTestPool()
			owner := uuid.New()
			ctx := ctxutil.WithUser(t.Context(), owner)

			// Set correct owner for all expenditures
			for _, e := range tc.initialExpenditures {
				e.Owner = owner
				e.Date = time.Now()
			}

			require.NoError(t, database.PersistExpenditures(ctx, pool, tc.initialExpenditures))

			// Create initial budget with existing expenses
			budget := model.NewBudget(owner, "test budget")
			for _, category := range tc.existingExpenses {
				budget.SetBasicExpense(category, 1000)
			}

			require.NoError(t, database.PersistBudget(ctx, pool, budget))

			// Update budget with budget
			for _, category := range tc.newExpenses {
				budget.SetBasicExpense(category, 1000)
			}

			require.NoError(t, database.PersistBudget(ctx, pool, budget))

			// Verify classifications
			fetched, err := database.ListExpenditures(ctx, pool, nil, nil, nil, nil,
				time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1), nil, nil)
			require.NoError(t, err)

			// Get the new expense IDs from the budget
			expenseIDs := make(map[string]uuid.UUID)
			budgets, err := database.GetBudgets(ctx, pool, owner, 1)
			require.NoError(t, err)
			require.Len(t, budgets, 1)

			for _, expense := range budgets[0].Expenses {
				expenseIDs[expense.Category] = expense.ID
			}

			// Count classified expenditures by category
			classified := make(map[string]int)

			for _, e := range fetched {
				if e.ExpenseID == expenseIDs[e.BudgetCategory] {
					classified[e.BudgetCategory]++
				}
			}

			require.Equal(t, tc.expectedMatches, classified)
		})
	}
}

func findExpenseByCategory(expenses []*model.Expense, category string) *model.Expense {
	for _, e := range expenses {
		if e.Category == category {
			return e
		}
	}

	return nil
}
