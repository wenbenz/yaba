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
	require.Equal(t, budget1, budget2)
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
	startDate, _ := time.ParseInLocation(time.DateOnly, startDateString, time.UTC)
	endDate, _ := time.ParseInLocation(time.DateOnly, endDateString, time.UTC)
	limit := 301

	err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
	require.NoError(t, err)

	persistedExpenditures, err := database.ListExpenditures(ctx, pool, startDate, endDate, nil, nil, &limit, nil)
	require.NoError(t, err)
	require.Len(t, persistedExpenditures, 300)

	expenditures, err := resolver.Query().Expenditures(ctx, &startDateString, &endDateString, nil, nil, &limit, nil)
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
	expenditures, err = resolver.Query().Expenditures(ctx, nil, nil, nil, nil, nil, nil)
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
	startDate, _ := time.ParseInLocation(time.DateOnly, startDateString, time.UTC)
	endDate, _ := time.ParseInLocation(time.DateOnly, endDateString, time.UTC)

	err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
	require.NoError(t, err)

	aggregate, err := resolver.Query().AggregatedExpenditures(ctx, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, aggregate, 33)
}

func TestCreateExpenditures(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		inputs := []*model.ExpenditureInput{
			{
				Name:           ptr("Expense 1"),
				Amount:         100.50,
				Date:           "2024-03-20",
				Method:         ptr("credit"),
				BudgetCategory: ptr("groceries"),
				Comment:        ptr("Test expense 1"),
			},
			{
				Name:           ptr("Expense 2"),
				Amount:         50.25,
				Date:           "2024-03-21",
				Method:         ptr("debit"),
				BudgetCategory: ptr("entertainment"),
				Comment:        ptr("Test expense 2"),
			},
		}

		success, err := resolver.Mutation().CreateExpenditures(ctx, inputs)
		require.NoError(t, err)
		require.True(t, *success)

		startDate, _ := time.ParseInLocation(time.DateOnly, "2024-03-20", time.UTC)
		endDate, _ := time.ParseInLocation(time.DateOnly, "2024-03-21", time.UTC)
		limit := 10

		expenditures, err := database.ListExpenditures(ctx, pool, startDate, endDate, nil, nil, &limit, nil)
		require.NoError(t, err)
		require.Len(t, expenditures, 2)

		// Verify second expenditure (comes first due to descending order)
		require.Equal(t, *inputs[1].Name, expenditures[0].Name)
		require.InDelta(t, inputs[1].Amount, expenditures[0].Amount, 0.001)
		require.Equal(t, *inputs[1].BudgetCategory, expenditures[0].BudgetCategory)
		require.Equal(t, *inputs[1].Method, expenditures[0].Method)
		require.Equal(t, *inputs[1].Comment, expenditures[0].Comment)

		// Verify first expenditure (comes second due to descending order)
		require.Equal(t, *inputs[0].Name, expenditures[1].Name)
		require.InDelta(t, inputs[0].Amount, expenditures[1].Amount, 0.001)
		require.Equal(t, *inputs[0].BudgetCategory, expenditures[1].BudgetCategory)
		require.Equal(t, *inputs[0].Method, expenditures[1].Method)
		require.Equal(t, *inputs[0].Comment, expenditures[1].Comment)
	})

	t.Run("invalid date format", func(t *testing.T) {
		t.Parallel()

		inputs := []*model.ExpenditureInput{{
			Name:           ptr("Invalid Date"),
			Amount:         100.50,
			Date:           "03-20-2024", // Wrong format
			Method:         ptr("credit"),
			BudgetCategory: ptr("groceries"),
		}}

		success, err := resolver.Mutation().CreateExpenditures(ctx, inputs)
		require.Error(t, err)
		require.False(t, *success)
	})

	t.Run("empty input array", func(t *testing.T) {
		t.Parallel()

		success, err := resolver.Mutation().CreateExpenditures(ctx, []*model.ExpenditureInput{})
		require.NoError(t, err)
		require.True(t, *success)
	})

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()

		inputs := []*model.ExpenditureInput{{
			Name:   ptr("Missing Fields"),
			Amount: 100.50,
			// Missing Date
			Method:         ptr("credit"),
			BudgetCategory: ptr("groceries"),
		}}

		success, err := resolver.Mutation().CreateExpenditures(ctx, inputs)
		require.Error(t, err)
		require.False(t, *success)
	})
}

//nolint:paralleltest
func TestCreateRewardCard(t *testing.T) {
	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.NewIsolatedTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	input := model.RewardCardInput{
		Name:   "Chase Freedom Flex",
		Issuer: "Chase",
		Region: "US",

		RewardType: "points",
	}

	// Create the reward card
	result, err := resolver.Mutation().CreateRewardCard(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the created reward card fields
	require.Equal(t, input.Name, result.Name)
	require.Equal(t, input.Issuer, result.Issuer)
	require.Equal(t, input.Region, result.Region)

	require.Equal(t, input.RewardType, result.RewardType)

	// Verify persistence using RewardCards query
	cards, err := resolver.Query().RewardCards(ctx, &input.Issuer, &input.Name, &input.Region, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 1)

	// Verify the queried card matches the created one
	require.Equal(t, result.ID, cards[0].ID)
	require.Equal(t, result.Name, cards[0].Name)
	require.Equal(t, result.Issuer, cards[0].Issuer)
	require.Equal(t, result.Region, cards[0].Region)
	require.Equal(t, result.RewardType, cards[0].RewardType)
}

func TestCreatePaymentMethodSuccess(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	// Create a reward card first
	rewardCard, err := resolver.Mutation().CreateRewardCard(ctx, model.RewardCardInput{
		Name:   "Chase Sapphire Reserve",
		Issuer: "Chase",
		Region: "US",

		RewardType: "points",
	})
	require.NoError(t, err)
	require.NotNil(t, rewardCard)

	// Create the payment method
	input := model.PaymentMethodInput{
		DisplayName:  ptr("Test Credit Card"),
		CardType:     ptr(rewardCard.ID),
		AcquiredDate: ptr("2024-03-20"),
		CancelByDate: ptr("2025-03-20"),
	}

	result, err := resolver.Mutation().CreatePaymentMethod(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the created payment method
	require.Equal(t, *input.DisplayName, result.DisplayName)
	require.Equal(t, *input.CardType, result.CardType)
	require.Equal(t, *input.AcquiredDate, *result.AcquiredDate)
	require.Equal(t, *input.CancelByDate, *result.CancelByDate)

	// Verify it was persisted
	persisted, err := database.GetPaymentMethod(ctx, pool, uuid.MustParse(result.ID))
	require.NoError(t, err)
	require.Equal(t, *input.DisplayName, persisted.DisplayName)
	require.Equal(t, *input.CardType, persisted.CardType.String())
}

func TestCreatePaymentMethodMissingFields(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	input := model.PaymentMethodInput{
		DisplayName: ptr("Test Card"),
		// Missing other required fields
	}

	result, err := resolver.Mutation().CreatePaymentMethod(ctx, input)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestUpdatePaymentMethod(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	// Create a reward card first
	rewardCard, err := resolver.Mutation().CreateRewardCard(ctx, model.RewardCardInput{
		Name:   "Chase Sapphire Reserve",
		Issuer: "Chase",
		Region: "US",

		RewardType: "points",
	})
	require.NoError(t, err)
	require.NotNil(t, rewardCard)

	// Create another reward card for updating
	newRewardCard, err := resolver.Mutation().CreateRewardCard(ctx, model.RewardCardInput{
		Name:   "Amex Gold",
		Issuer: "American Express",
		Region: "US",

		RewardType: "points",
	})
	require.NoError(t, err)
	require.NotNil(t, newRewardCard)

	// Create initial payment method
	createInput := model.PaymentMethodInput{
		DisplayName:  ptr("Original Card"),
		CardType:     ptr(rewardCard.ID),
		AcquiredDate: ptr("2024-03-20"),
		CancelByDate: ptr("2025-03-20"),
	}

	created, err := resolver.Mutation().CreatePaymentMethod(ctx, createInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Update the payment method
	updateInput := model.PaymentMethodInput{
		DisplayName:  ptr("Updated Card"),
		CardType:     ptr(newRewardCard.ID),
		AcquiredDate: ptr("2024-03-21"),
		CancelByDate: ptr("2025-03-21"),
	}

	updated, err := resolver.Mutation().UpdatePaymentMethod(ctx, created.ID, updateInput)
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Verify the updated fields
	require.Equal(t, *updateInput.DisplayName, updated.DisplayName)
	require.Equal(t, *updateInput.CardType, updated.CardType)
	require.Equal(t, *updateInput.AcquiredDate, *updated.AcquiredDate)
	require.Equal(t, *updateInput.CancelByDate, *updated.CancelByDate)

	// Verify persistence
	persisted, err := database.GetPaymentMethod(ctx, pool, uuid.MustParse(updated.ID))
	require.NoError(t, err)
	require.Equal(t, *updateInput.DisplayName, persisted.DisplayName)
	require.Equal(t, *updateInput.CardType, persisted.CardType.String())

	// Verify using the query resolver
	methods, err := resolver.Query().PaymentMethods(ctx)
	require.NoError(t, err)
	require.Len(t, methods, 1)
	require.Equal(t, updated.ID, methods[0].ID)
	require.Equal(t, updated.DisplayName, methods[0].DisplayName)
}

func TestDeletePaymentMethod(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	// Create a reward card first
	rewardCard, err := resolver.Mutation().CreateRewardCard(ctx, model.RewardCardInput{
		Name:   "Chase Sapphire Reserve",
		Issuer: "Chase",
		Region: "US",

		RewardType: "points",
	})
	require.NoError(t, err)
	require.NotNil(t, rewardCard)

	// Create a payment method to delete
	input := model.PaymentMethodInput{
		DisplayName:  ptr("Test Credit Card"),
		CardType:     ptr(rewardCard.ID),
		AcquiredDate: ptr("2024-03-20"),
		CancelByDate: ptr("2025-03-20"),
	}

	created, err := resolver.Mutation().CreatePaymentMethod(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Verify payment method exists
	methods, err := resolver.Query().PaymentMethods(ctx)
	require.NoError(t, err)
	require.Len(t, methods, 1)

	// Delete the payment method
	success, err := resolver.Mutation().DeletePaymentMethod(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, success)

	// Verify payment method was deleted
	methods, err = resolver.Query().PaymentMethods(ctx)
	require.NoError(t, err)
	require.Empty(t, methods)

	// Try to delete non-existent payment method
	success, err = resolver.Mutation().DeletePaymentMethod(ctx, created.ID)
	require.NoError(t, err)
	require.False(t, success)
}

func TestPaymentMethods_Empty(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	methods, err := resolver.Query().PaymentMethods(ctx)
	require.NoError(t, err)
	require.Empty(t, methods)
}

func TestPaymentMethods_WithCards(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.GetTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	// Create a reward card first
	rewardCard, err := resolver.Mutation().CreateRewardCard(ctx, model.RewardCardInput{
		Name:   "Chase Sapphire Reserve",
		Issuer: "Chase",
		Region: "US",

		RewardType: "points",
	})
	require.NoError(t, err)
	require.NotNil(t, rewardCard)

	// Create multiple payment methods
	inputs := []model.PaymentMethodInput{
		{
			DisplayName:  ptr("Card 1"),
			CardType:     ptr(rewardCard.ID),
			AcquiredDate: ptr("2024-03-20"),
			CancelByDate: ptr("2025-03-20"),
		},
		{
			DisplayName:  ptr("Card 2"),
			CardType:     ptr(rewardCard.ID),
			AcquiredDate: ptr("2024-03-21"),
			CancelByDate: ptr("2025-03-21"),
		},
	}

	for _, input := range inputs {
		result, err := resolver.Mutation().CreatePaymentMethod(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, *input.DisplayName, result.DisplayName)
	}

	// Create a payment method under a different user
	otherUser := uuid.New()
	otherCtx := ctxutil.WithUser(t.Context(), otherUser)
	result, err := resolver.Mutation().CreatePaymentMethod(otherCtx,
		model.PaymentMethodInput{CardType: ptr(rewardCard.ID)})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEqual(t, result.ID, uuid.Nil)

	// Query all payment methods
	methods, err := resolver.Query().PaymentMethods(ctx)
	require.NoError(t, err)
	require.Len(t, methods, 2)

	// Verify the payment methods match the inputs
	for i, method := range methods {
		require.Equal(t, *inputs[i].DisplayName, method.DisplayName)
		require.Equal(t, *inputs[i].CardType, method.CardType)
		require.Equal(t, *inputs[i].AcquiredDate, *method.AcquiredDate)
		require.Equal(t, *inputs[i].CancelByDate, *method.CancelByDate)
	}
}

//nolint:paralleltest
func TestRewardCards_empty(t *testing.T) {
	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.NewIsolatedTestPool()
	resolver := &handlers.Resolver{Pool: pool}

	cards, err := resolver.Query().RewardCards(ctx, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Empty(t, cards)
}

//nolint:paralleltest
func TestRewardCards(t *testing.T) {
	user := uuid.New()
	ctx := ctxutil.WithUser(t.Context(), user)
	pool := helper.NewIsolatedTestPool()
	resolver := &handlers.Resolver{Pool: pool}
	// Create multiple reward cards
	inputs := []model.RewardCardInput{
		{
			Name:   "Chase Sapphire Reserve",
			Issuer: "Chase",
			Region: "US",

			RewardType: "points",
		},
		{
			Name:   "Amex Gold",
			Issuer: "American Express",
			Region: "US",

			RewardType: "points",
		},
		{
			Name:   "Chase Freedom Flex",
			Issuer: "Chase",
			Region: "US",

			RewardType: "points",
		},
	}

	// Create reward cards
	for _, input := range inputs {
		result, err := resolver.Mutation().CreateRewardCard(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, input.Name, result.Name)
	}

	// Test no filters
	cards, err := resolver.Query().RewardCards(ctx, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Test issuer filter
	issuer := "Chase"
	cards, err = resolver.Query().RewardCards(ctx, &issuer, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 2)

	for _, card := range cards {
		require.Equal(t, "Chase", card.Issuer)
	}

	// Test name filter
	name := "Amex Gold"
	cards, err = resolver.Query().RewardCards(ctx, nil, &name, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Equal(t, "Amex Gold", cards[0].Name)

	// Test region filter
	region := "US"
	cards, err = resolver.Query().RewardCards(ctx, nil, nil, &region, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 3)

	for _, card := range cards {
		require.Equal(t, "US", card.Region)
	}

	// Test combined filters
	cards, err = resolver.Query().RewardCards(ctx, &issuer, nil, &region, nil, nil)
	require.NoError(t, err)
	require.Len(t, cards, 2)

	for _, card := range cards {
		require.Equal(t, "Chase", card.Issuer)
		require.Equal(t, "US", card.Region)
	}
}

func ptr(s string) *string {
	return &s
}
