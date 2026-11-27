package rpc

import (
	"context"
	"errors"
	"math"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	yabaerrors "yaba/errors"
	yabav1 "yaba/gen/yaba/v1"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
)

var errBudgetNotFound = errors.New("budget not found")

// BudgetServer implements BudgetService for the user in the request context.
type BudgetServer struct {
	Pool *pgxpool.Pool
}

func (s *BudgetServer) GetBudget(
	ctx context.Context,
	req *connect.Request[yabav1.GetBudgetRequest],
) (*connect.Response[yabav1.GetBudgetResponse], error) {
	budget, err := s.getBudget(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&yabav1.GetBudgetResponse{Budget: budgetToProto(budget)}), nil
}

func (s *BudgetServer) ListBudgets(
	ctx context.Context,
	req *connect.Request[yabav1.ListBudgetsRequest],
) (*connect.Response[yabav1.ListBudgetsResponse], error) {
	limit := 10

	if req.Msg.Limit != nil {
		limit = int(req.Msg.GetLimit())
	}

	budgets, err := database.GetBudgets(ctx, s.Pool, ctxutil.GetUser(ctx), limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.Budget, 0, len(budgets))
	for _, budget := range budgets {
		out = append(out, budgetToProto(budget))
	}

	return connect.NewResponse(&yabav1.ListBudgetsResponse{Budgets: out}), nil
}

func (s *BudgetServer) CreateBudget(
	ctx context.Context,
	req *connect.Request[yabav1.CreateBudgetRequest],
) (*connect.Response[yabav1.CreateBudgetResponse], error) {
	msg := req.Msg
	id := uuid.New()
	budget := &model.Budget{
		ID:       id,
		Name:     msg.GetName(),
		Incomes:  incomesFromProto(id, msg.GetIncomes()),
		Expenses: expensesFromProto(id, msg.GetExpenses()),
	}

	err := database.PersistBudget(ctx, s.Pool, budget)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
		pgErr.Code == "23505" && pgErr.ConstraintName == "idx_owner_name" {
		// Budget with this name already exists; return it (idempotent create).
		budget, err = database.GetBudgetByName(ctx, s.Pool, ctxutil.GetUser(ctx), msg.GetName())
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.CreateBudgetResponse{Budget: budgetToProto(budget)}), nil
}

func (s *BudgetServer) UpdateBudget(
	ctx context.Context,
	req *connect.Request[yabav1.UpdateBudgetRequest],
) (*connect.Response[yabav1.UpdateBudgetResponse], error) {
	in := req.Msg.GetBudget()

	budget, err := s.getBudget(ctx, in.GetId())
	if err != nil {
		return nil, err
	}

	for _, path := range req.Msg.GetUpdateMask().GetPaths() {
		switch path {
		case "name":
			budget.Name = in.GetName()
		case "incomes":
			budget.Incomes = incomesFromProto(budget.ID, in.GetIncomes())
		case "expenses":
			budget.Expenses = expensesFromProto(budget.ID, in.GetExpenses())
		}
	}

	if err = database.PersistBudget(ctx, s.Pool, budget); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.UpdateBudgetResponse{Budget: budgetToProto(budget)}), nil
}

func (s *BudgetServer) getBudget(ctx context.Context, id string) (*model.Budget, error) {
	// The id is validated by protovalidate.
	budget, err := database.GetBudget(ctx, s.Pool, ctxutil.GetUser(ctx), uuid.MustParse(id))
	if _, ok := errors.AsType[yabaerrors.NoSuchElementError](err); ok {
		return nil, connect.NewError(connect.CodeNotFound, errBudgetNotFound)
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return budget, nil
}

func budgetToProto(budget *model.Budget) *yabav1.Budget {
	out := &yabav1.Budget{Id: budget.ID.String(), Name: budget.Name}

	for _, income := range budget.Incomes {
		out.Incomes = append(out.Incomes, &yabav1.Income{
			Source: income.Source,
			Amount: int64(math.Round(income.Amount * 100)),
		})
	}

	for _, expense := range budget.Expenses {
		out.Expenses = append(out.Expenses, &yabav1.Expense{
			Id:       expense.ID.String(),
			Category: expense.Category,
			Amount:   int64(math.Round(expense.Amount * 100)),
			IsFixed:  expense.Fixed,
			IsSlack:  expense.Slack,
		})
	}

	return out
}

func incomesFromProto(budgetID uuid.UUID, incomes []*yabav1.Income) []*model.Income {
	out := make([]*model.Income, 0, len(incomes))
	for _, income := range incomes {
		out = append(out, &model.Income{
			Owner:  budgetID,
			Source: income.GetSource(),
			Amount: float64(income.GetAmount()) / 100,
		})
	}

	return out
}

func expensesFromProto(budgetID uuid.UUID, expenses []*yabav1.Expense) []*model.Expense {
	out := make([]*model.Expense, 0, len(expenses))
	for _, expense := range expenses {
		id, _ := uuid.Parse(expense.GetId()) // validated by protovalidate; empty parses to uuid.Nil
		out = append(out, &model.Expense{
			BudgetID: budgetID,
			ID:       id,
			Category: expense.GetCategory(),
			Amount:   float64(expense.GetAmount()) / 100,
			Fixed:    expense.GetIsFixed(),
			Slack:    expense.GetIsSlack(),
		})
	}

	return out
}
