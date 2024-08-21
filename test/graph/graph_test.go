package graph_test

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
	"yaba/internal/database"
	"yaba/internal/graph/client"
	"yaba/internal/server"
	"yaba/test/helper"
)

func TestExpenditures(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	ctx := context.Background()
	pool := helper.GetTestPool()
	svr := httptest.NewServer(server.BuildServerHandler(pool))

	startDate, _ := time.Parse(time.DateOnly, "2020-01-01")
	endDate, _ := time.Parse(time.DateOnly, "2020-02-01")

	err := database.PersistExpenditures(ctx, pool, helper.MockExpenditures(300, user, startDate, endDate))
	require.NoError(t, err)

	persistedExpenditures, err := database.ListExpenditures(ctx, pool, user, startDate, endDate, 301)
	require.NoError(t, err)
	require.Len(t, persistedExpenditures, 300)

	gql := graphql.NewClient(svr.URL+"/graphql", svr.Client())
	expenditures, err := client.ListExpenditures(ctx, gql,
		startDate.Format(time.DateOnly), endDate.Format(time.DateOnly), 302)
	require.NoError(t, err)
	require.Len(t, expenditures.GetExpenditures(), 300)

	for i, expenditure := range expenditures.GetExpenditures() {
		require.Equal(t, persistedExpenditures[i].Name, expenditure.GetName())
		amount, err := strconv.ParseFloat(expenditure.GetAmount(), 64)
		require.NoError(t, err)
		require.InDelta(t, persistedExpenditures[i].Amount, amount, .009)
		require.Equal(t, persistedExpenditures[i].CreatedTime.Format(time.DateOnly), expenditure.GetCreated())
		require.Equal(t, persistedExpenditures[i].Date.Format(time.DateOnly), expenditure.GetDate())
		require.Equal(t, persistedExpenditures[i].Comment, expenditure.GetComment())
		require.Equal(t, persistedExpenditures[i].BudgetCategory, expenditure.GetBudget_category())
		require.Equal(t, strconv.Itoa(persistedExpenditures[i].ID), expenditure.GetId())
		require.Equal(t, persistedExpenditures[i].Method, expenditure.GetMethod())
		require.Equal(t, persistedExpenditures[i].Source, expenditure.GetSource())

		if persistedExpenditures[i].RewardCategory.Valid {
			require.Equal(t, persistedExpenditures[i].RewardCategory.String, expenditure.GetReward_category())
		} else {
			require.Equal(t, "", expenditure.GetReward_category())
		}
	}
}

func TestCreateEmptyBudget(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	ctx := context.Background()
	pool := helper.GetTestPool()
	svr := httptest.NewServer(server.BuildServerHandler(pool))

	gql := graphql.NewClient(svr.URL+"/graphql", svr.Client())
	listBudgetsResponse, err := client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Empty(t, listBudgetsResponse.Budgets)

	_, err = client.CreateBudget(ctx, gql, "Budget V1", nil, nil)
	require.NoError(t, err)

	listBudgetsResponse, err = client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Len(t, listBudgetsResponse.Budgets, 1)

	budget1 := listBudgetsResponse.Budgets[0]

	require.NoError(t, err)
	require.NotNil(t, budget1.GetId())
	require.Equal(t, "Budget V1", budget1.GetName())
	require.Empty(t, budget1.GetIncomes())
	require.Empty(t, budget1.GetExpenses())

	getBudgetsResponse, err := client.GetBudget(ctx, gql, budget1.GetId())
	require.NoError(t, err)

	budget2 := getBudgetsResponse.GetBudget()
	require.EqualValues(t, budget1.GetId(), budget2.GetId())
}

func TestCreateFullBudget(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	ctx := context.Background()
	pool := helper.GetTestPool()
	svr := httptest.NewServer(server.BuildServerHandler(pool))

	gql := graphql.NewClient(svr.URL+"/graphql", svr.Client())
	listBudgetsResponse, err := client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Empty(t, listBudgetsResponse.Budgets)

	response, err := client.CreateBudget(ctx, gql, "Budget2", []client.IncomeInput{
		{
			Source: "work",
			Amount: 100_000.00,
		},
	}, []client.ExpenseInput{
		{
			Category: "rent",
			Amount:   2_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "food",
			Amount:   1_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "entertainment",
			Amount:   1_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "savings",
			Amount:   1_000,
			IsFixed:  false,
			IsSlack:  true,
		},
	})
	require.NoError(t, err)

	b := response.GetCreateBudget()
	require.NotNil(t, b.GetId())
	require.Equal(t, "Budget2", b.GetName())
	require.Equal(t, b.GetOwner(), user.String())
	require.Len(t, b.GetIncomes(), 1)
	require.Len(t, b.GetExpenses(), 4)

	listBudgetsResponse, err = client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Len(t, listBudgetsResponse.Budgets, 1)
}

func TestUpdateBudget(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	ctx := context.Background()
	pool := helper.GetTestPool()
	svr := httptest.NewServer(server.BuildServerHandler(pool))

	gql := graphql.NewClient(svr.URL+"/graphql", svr.Client())
	listBudgetsResponse, err := client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Empty(t, listBudgetsResponse.Budgets)

	// Create the budget
	_, err = client.CreateBudget(ctx, gql, "Budget1", []client.IncomeInput{
		{
			Source: "work",
			Amount: 100_000.00,
		},
	}, []client.ExpenseInput{
		{
			Category: "rent",
			Amount:   2_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "food",
			Amount:   1_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "entertainment",
			Amount:   1_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "savings",
			Amount:   1_000,
			IsFixed:  false,
			IsSlack:  true,
		},
	})
	require.NoError(t, err)

	// Check that the budget is what we set it to.
	listBudgetsResponse, err = client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Len(t, listBudgetsResponse.Budgets, 1)
	b := listBudgetsResponse.Budgets[0]
	require.NotNil(t, b.GetId())
	require.Equal(t, "Budget1", b.GetName())
	require.Len(t, b.GetIncomes(), 1)
	require.Len(t, b.GetExpenses(), 4)

	// Update the budget
	_, err = client.UpdateBudget(ctx, gql, b.GetId(), "Budget2", []client.IncomeInput{
		{
			Source: "work",
			Amount: 100_000.00,
		},
		{
			Source: "uber",
			Amount: 20_000,
		},
	}, []client.ExpenseInput{
		{
			Category: "rent",
			Amount:   2_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "food",
			Amount:   1_000,
			IsFixed:  true,
			IsSlack:  false,
		}, {
			Category: "savings",
			Amount:   1_000,
			IsFixed:  false,
			IsSlack:  true,
		},
	})
	require.NoError(t, err)

	// Check that the budget has been updated
	listBudgetsResponse, err = client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Len(t, listBudgetsResponse.Budgets, 1)

	b = listBudgetsResponse.Budgets[0]
	require.NotNil(t, b.GetId())
	require.Equal(t, "Budget2", b.GetName())
	require.Len(t, b.GetIncomes(), 2)
	require.Len(t, b.GetExpenses(), 3)
}

func TestUpdateFailsWrongOwner(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	ctx := context.Background()
	pool := helper.GetTestPool()
	svr := httptest.NewServer(server.BuildServerHandler(pool))

	gql := graphql.NewClient(svr.URL+"/graphql", svr.Client())
	listBudgetsResponse, err := client.ListBudgets(ctx, gql, 10)
	require.NoError(t, err)
	require.Empty(t, listBudgetsResponse.Budgets)

	response, err := client.CreateBudget(ctx, gql, "Budget V1", nil, nil)
	require.NoError(t, err)

	// Just gonna go in and change the owner
	tag, err := pool.Exec(ctx, "UPDATE budget SET owner = $2 WHERE owner = $1;", user, uuid.New())
	require.NoError(t, err)
	require.Equal(t, int64(1), tag.RowsAffected())

	b := response.GetCreateBudget()
	_, err = client.UpdateBudget(ctx, gql, b.GetId(), "Budget V2", nil, nil)
	require.ErrorContains(t, err, "user does not own this budget")

	budgetID, err := uuid.Parse(b.GetId())
	require.NoError(t, err)

	dbBudget, err := database.GetBudget(ctx, pool, budgetID)
	require.NoError(t, err)
	require.Equal(t, budgetID, dbBudget.ID)
	require.Equal(t, "Budget V1", dbBudget.Name)
}
