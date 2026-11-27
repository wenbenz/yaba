package rpc_test

import (
	"testing"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func newBudgetClient(t *testing.T, pool *pgxpool.Pool, user uuid.UUID) yabav1connect.BudgetServiceClient {
	t.Helper()

	srv := serve(t, pool, user)

	return yabav1connect.NewBudgetServiceClient(srv.Client(), srv.URL+"/rpc")
}

func createBudget(
	t *testing.T,
	client yabav1connect.BudgetServiceClient,
	req *yabav1.CreateBudgetRequest,
) *yabav1.Budget {
	t.Helper()

	res, err := client.CreateBudget(t.Context(), connect.NewRequest(req))
	require.NoError(t, err)

	return res.Msg.GetBudget()
}

func listBudgets(
	t *testing.T,
	client yabav1connect.BudgetServiceClient,
	req *yabav1.ListBudgetsRequest,
) []*yabav1.Budget {
	t.Helper()

	res, err := client.ListBudgets(t.Context(), connect.NewRequest(req))
	require.NoError(t, err)

	return res.Msg.GetBudgets()
}

func TestCreateGetAndListBudgets(t *testing.T) {
	t.Parallel()

	client := newBudgetClient(t, helper.GetTestPool(), uuid.New())
	req := &yabav1.CreateBudgetRequest{
		Name:     "Household",
		Incomes:  []*yabav1.Income{{Source: "Salary", Amount: 500000}},
		Expenses: []*yabav1.Expense{{Category: "Groceries", Amount: 12345, IsFixed: true}},
	}

	created := createBudget(t, client, req)
	require.NoError(t, uuid.Validate(created.GetId()))
	require.Len(t, created.GetExpenses(), 1)

	expenseID := created.GetExpenses()[0].GetId()
	require.NoError(t, uuid.Validate(expenseID))
	require.NotEqual(t, uuid.Nil.String(), expenseID)

	want := &yabav1.Budget{
		Id:       created.GetId(),
		Name:     "Household",
		Incomes:  req.GetIncomes(),
		Expenses: []*yabav1.Expense{{Id: expenseID, Category: "Groceries", Amount: 12345, IsFixed: true}},
	}
	require.True(t, proto.Equal(want, created), "created: %v", created)

	getReq := connect.NewRequest(&yabav1.GetBudgetRequest{Id: created.GetId()})
	got, err := client.GetBudget(t.Context(), getReq)
	require.NoError(t, err)
	require.True(t, proto.Equal(want, got.Msg.GetBudget()), "got: %v", got.Msg.GetBudget())

	listed := listBudgets(t, client, &yabav1.ListBudgetsRequest{})
	require.Len(t, listed, 1)
	require.True(t, proto.Equal(want, listed[0]), "listed: %v", listed[0])

	again := createBudget(t, client, &yabav1.CreateBudgetRequest{Name: "Household"})
	require.True(t, proto.Equal(want, again), "same name returns the existing budget: %v", again)

	createBudget(t, client, &yabav1.CreateBudgetRequest{Name: "Vacation"})
	require.Len(t, listBudgets(t, client, &yabav1.ListBudgetsRequest{}), 2)
	require.Len(t, listBudgets(t, client, &yabav1.ListBudgetsRequest{Limit: new(int32(1))}), 1)
	require.Empty(t, listBudgets(t, client, &yabav1.ListBudgetsRequest{Limit: new(int32(0))}))

	other := newBudgetClient(t, helper.GetTestPool(), uuid.New())
	require.Empty(t, listBudgets(t, other, &yabav1.ListBudgetsRequest{}), "other users must not see these")

	_, err = other.GetBudget(t.Context(), getReq)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err), "other users can't get it")
}

func TestUpdateBudget(t *testing.T) {
	t.Parallel()

	client := newBudgetClient(t, helper.GetTestPool(), uuid.New())
	created := createBudget(t, client, &yabav1.CreateBudgetRequest{
		Name:     "Household",
		Incomes:  []*yabav1.Income{{Source: "Salary", Amount: 500000}},
		Expenses: []*yabav1.Expense{{Category: "Groceries", Amount: 12345}},
	})

	// Only masked fields change, so the income sent here is ignored.
	rename := &yabav1.UpdateBudgetRequest{
		Budget: &yabav1.Budget{
			Id:      created.GetId(),
			Name:    "Renamed",
			Incomes: []*yabav1.Income{{Source: "Ignored", Amount: 1}},
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	}

	other := newBudgetClient(t, helper.GetTestPool(), uuid.New())
	_, err := other.UpdateBudget(t.Context(), connect.NewRequest(rename))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err), "other users can't update it")

	res, err := client.UpdateBudget(t.Context(), connect.NewRequest(rename))
	require.NoError(t, err)

	want := &yabav1.Budget{
		Id:       created.GetId(),
		Name:     "Renamed",
		Incomes:  created.GetIncomes(),
		Expenses: created.GetExpenses(),
	}
	require.True(t, proto.Equal(want, res.Msg.GetBudget()), "renamed: %v", res.Msg.GetBudget())

	replace := &yabav1.UpdateBudgetRequest{
		Budget: &yabav1.Budget{
			Id:       created.GetId(),
			Name:     "Ignored",
			Incomes:  []*yabav1.Income{{Source: "Bonus", Amount: 100000}},
			Expenses: []*yabav1.Expense{{Id: created.GetExpenses()[0].GetId(), Category: "Groceries", Amount: 20000}},
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"incomes", "expenses"}},
	}

	res, err = client.UpdateBudget(t.Context(), connect.NewRequest(replace))
	require.NoError(t, err)

	want = &yabav1.Budget{
		Id:       created.GetId(),
		Name:     "Renamed",
		Incomes:  replace.GetBudget().GetIncomes(),
		Expenses: replace.GetBudget().GetExpenses(),
	}
	require.True(t, proto.Equal(want, res.Msg.GetBudget()), "replaced: %v", res.Msg.GetBudget())

	got, err := client.GetBudget(t.Context(), connect.NewRequest(&yabav1.GetBudgetRequest{Id: created.GetId()}))
	require.NoError(t, err)
	require.True(t, proto.Equal(want, got.Msg.GetBudget()), "got: %v", got.Msg.GetBudget())
}

func TestBudgetService_InvalidArgument(t *testing.T) {
	t.Parallel()

	client := newBudgetClient(t, helper.GetTestPool(), uuid.New())

	_, err := client.CreateBudget(t.Context(), connect.NewRequest(&yabav1.CreateBudgetRequest{}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "empty name")

	_, err = client.CreateBudget(t.Context(), connect.NewRequest(&yabav1.CreateBudgetRequest{
		Name:     "Household",
		Expenses: []*yabav1.Expense{{Id: "not-a-uuid", Category: "Groceries"}},
	}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "malformed expense id")

	_, err = client.GetBudget(t.Context(), connect.NewRequest(&yabav1.GetBudgetRequest{Id: "not-a-uuid"}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "malformed id")

	_, err = client.ListBudgets(t.Context(), connect.NewRequest(&yabav1.ListBudgetsRequest{Limit: new(int32(-1))}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "negative limit")

	_, err = client.UpdateBudget(t.Context(), connect.NewRequest(&yabav1.UpdateBudgetRequest{
		Budget:     &yabav1.Budget{Id: uuid.NewString()},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"owner"}},
	}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "unknown mask path")
}

func TestBudgetService_RequiresUser(t *testing.T) {
	t.Parallel()

	client := newBudgetClient(t, helper.GetTestPool(), uuid.Nil)

	_, err := client.ListBudgets(t.Context(), connect.NewRequest(&yabav1.ListBudgetsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))

	// Auth runs before validation.
	_, err = client.GetBudget(t.Context(), connect.NewRequest(&yabav1.GetBudgetRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestBudgetService_DatabaseErrorIsInternal(t *testing.T) {
	t.Parallel()

	pool, err := pgxpool.NewWithConfig(t.Context(), helper.GetTestPool().Config())
	require.NoError(t, err)
	pool.Close()

	client := newBudgetClient(t, pool, uuid.New())
	id := uuid.NewString()

	_, err = client.GetBudget(t.Context(), connect.NewRequest(&yabav1.GetBudgetRequest{Id: id}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.ListBudgets(t.Context(), connect.NewRequest(&yabav1.ListBudgetsRequest{}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.CreateBudget(t.Context(), connect.NewRequest(&yabav1.CreateBudgetRequest{Name: "Household"}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.UpdateBudget(t.Context(), connect.NewRequest(&yabav1.UpdateBudgetRequest{
		Budget:     &yabav1.Budget{Id: id},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))
}
