package handlers_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
	"yaba/graph/model"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/handlers"

	"yaba/internal/test/helper"
)

func TestCreateEmptyBudget(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	budgets, err := resolver.Query().Budgets(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, budgets)

	_, err = resolver.Mutation().CreateBudget(ctx, model.NewBudgetInput{
		Name:     "Budget V1",
		Incomes:  nil,
		Expenses: nil,
	})

	require.NoError(t, err)

	budgets, err = resolver.Query().Budgets(ctx, nil)
	require.NoError(t, err)
	require.Len(t, budgets, 1)

	budget1 := budgets[0]

	require.NoError(t, err)
	require.NotNil(t, *budget1.ID)
	require.Equal(t, "Budget V1", *budget1.Name)
	require.Empty(t, budget1.Incomes)
	require.Empty(t, budget1.Expenses)

	budget2, err := resolver.Query().Budget(ctx, *budget1.ID)
	require.NoError(t, err)
	require.EqualValues(t, budget1, budget2)
}

func TestCreateFullBudget(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}
	limit := 10

	budgets, err := resolver.Query().Budgets(ctx, &limit)
	require.NoError(t, err)
	require.Empty(t, budgets)

	b, err := resolver.Mutation().CreateBudget(ctx, model.NewBudgetInput{
		Name: "Budget2",
		Incomes: []*model.IncomeInput{
			{
				Source: "work",
				Amount: 100_000.00,
			},
		},
		Expenses: []*model.ExpenseInput{
			{
				Category: "rent",
				Amount:   2_000,
			}, {
				Category: "food",
				Amount:   1_000,
			}, {
				Category: "entertainment",
				Amount:   1_000,
			}, {
				Category: "savings",
				Amount:   1_000,
			},
		},
	})

	require.NoError(t, err)

	require.NotNil(t, *b.ID)
	require.Equal(t, "Budget2", *b.Name)
	require.Equal(t, user.String(), *b.Owner)
	require.Len(t, b.Incomes, 1)
	require.Len(t, b.Expenses, 4)

	budgets, err = resolver.Query().Budgets(ctx, &limit)
	require.NoError(t, err)
	require.Len(t, budgets, 1)
}

func TestUpdateBudget(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}
	limit := 10

	budgets, err := resolver.Query().Budgets(ctx, &limit)
	require.NoError(t, err)
	require.Empty(t, budgets)

	b, err := resolver.Mutation().CreateBudget(ctx, model.NewBudgetInput{
		Name: "Budget1",
		Incomes: []*model.IncomeInput{
			{
				Source: "work",
				Amount: 100_000.00,
			},
		},
		Expenses: []*model.ExpenseInput{
			{
				Category: "rent",
				Amount:   2_000,
			}, {
				Category: "food",
				Amount:   1_000,
			}, {
				Category: "entertainment",
				Amount:   1_000,
			}, {
				Category: "savings",
				Amount:   1_000,
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, *b.ID)
	require.Equal(t, "Budget1", *b.Name)
	require.Equal(t, user.String(), *b.Owner)
	require.Len(t, b.Incomes, 1)
	require.Len(t, b.Expenses, 4)

	newName := "Budget2"
	_, err = resolver.Mutation().UpdateBudget(ctx, model.UpdateBudgetInput{
		ID:   *b.ID,
		Name: &newName,
		Incomes: []*model.IncomeInput{
			{
				Source: "work",
				Amount: 100_000.00,
			},
			{
				Source: "uber",
				Amount: 20_000,
			},
		},
		Expenses: []*model.ExpenseInput{
			{
				Category: "rent",
				Amount:   2_000,
			}, {
				Category: "food",
				Amount:   1_000,
			}, {
				Category: "savings",
				Amount:   1_000,
			},
		},
	})

	require.NoError(t, err)

	// Check that the budget has been updated
	budgets, err = resolver.Query().Budgets(ctx, &limit)
	require.NoError(t, err)
	require.Len(t, budgets, 1)

	b = budgets[0]
	require.NotNil(t, *b.ID)
	require.Equal(t, "Budget2", *b.Name)
	require.Len(t, b.Incomes, 2)
	require.Len(t, b.Expenses, 3)
}

func TestUpdateFailsWrongOwner(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}
	limit := 10

	budgets, err := resolver.Query().Budgets(ctx, &limit)
	require.NoError(t, err)
	require.Empty(t, budgets)

	b, err := resolver.Mutation().CreateBudget(ctx, model.NewBudgetInput{
		Name:     "Budget V1",
		Incomes:  nil,
		Expenses: nil,
	})

	require.NoError(t, err)

	// Just gonna go in and change the owner
	newOwner := uuid.New()
	tag, err := pool.Exec(ctx, "UPDATE budget SET owner = $2 WHERE owner = $1;", user, newOwner)
	require.NoError(t, err)
	require.Equal(t, int64(1), tag.RowsAffected())

	newName := "randomblah"
	_, err = resolver.Mutation().UpdateBudget(ctx, model.UpdateBudgetInput{
		ID:   *b.ID,
		Name: &newName,
	})

	require.ErrorContains(t, err, "budget not found")

	budgetID, err := uuid.Parse(*b.ID)
	require.NoError(t, err)

	dbBudget, err := database.GetBudget(ctx, pool, newOwner, budgetID)
	require.NoError(t, err)
	require.Equal(t, budgetID, dbBudget.ID)
	require.Equal(t, "Budget V1", dbBudget.Name)
}

func TestExpenditures(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	startDateString, endDateString := "2020-01-01", "2020-02-01"
	startDate, _ := time.Parse(time.DateOnly, startDateString)
	endDate, _ := time.Parse(time.DateOnly, endDateString)
	limit := 301

	err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
	require.NoError(t, err)

	persistedExpenditures, err := database.ListExpenditures(ctx, pool, startDate, endDate, nil, limit)
	require.NoError(t, err)
	require.Len(t, persistedExpenditures, 300)

	expenditures, err := resolver.Query().Expenditures(ctx, &startDateString, &endDateString, nil, &limit)
	require.NoError(t, err)
	require.Len(t, expenditures, 300)

	for i, expenditure := range expenditures {
		require.Equal(t, persistedExpenditures[i].Name, *expenditure.Name)
		amount, err := strconv.ParseFloat(*expenditure.Amount, 64)
		require.NoError(t, err)
		require.InDelta(t, persistedExpenditures[i].Amount, amount, .009)
		require.Equal(t, persistedExpenditures[i].CreatedTime.Format(time.DateOnly), *expenditure.Created)
		require.Equal(t, persistedExpenditures[i].Date.Format(time.DateOnly), *expenditure.Date)
		require.Equal(t, persistedExpenditures[i].Comment, *expenditure.Comment)
		require.Equal(t, persistedExpenditures[i].BudgetCategory, *expenditure.BudgetCategory)
		require.Equal(t, strconv.Itoa(persistedExpenditures[i].ID), *expenditure.ID)
		require.Equal(t, persistedExpenditures[i].Method, *expenditure.Method)
		require.Equal(t, persistedExpenditures[i].Source, *expenditure.Source)
		require.Equal(t, persistedExpenditures[i].RewardCategory, *expenditure.RewardCategory)
	}

	// nil range should return everything
	expenditures, err = resolver.Query().Expenditures(ctx, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, expenditures, 10)
}

func TestAggregateExpenditures(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	startDateString, endDateString := "2020-01-01", "2020-02-02"
	startDate, _ := time.Parse(time.DateOnly, startDateString)
	endDate, _ := time.Parse(time.DateOnly, endDateString)

	err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
	require.NoError(t, err)

	aggregate, err := resolver.Query().AggregatedExpenditures(ctx, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, aggregate, 33)
}
