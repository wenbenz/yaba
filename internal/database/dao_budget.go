package database

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"yaba/errors"
	"yaba/internal/ctxutil"
	"yaba/internal/model"
)

const getBudget = `
SELECT * FROM budget
WHERE owner = $1
  AND id = $2;
`

const getBudgetsByOwner = `
SELECT * FROM budget
WHERE owner = $1
LIMIT $2;
`

const deleteBudget = `
DELETE FROM budget
WHERE owner = $1
  AND id = $2;
`

const getIncomesByOwner = `
SELECT * FROM income
WHERE owner IN ($1)
`

const upsertIncome = `
INSERT INTO income (owner, source, amount)
VALUES ($1, $2, $3)
ON CONFLICT (owner, source) DO UPDATE
SET amount = $3
`

const deleteIncomeByOwner = `
DELETE FROM income
WHERE owner = $1
`

const getExpensesForBudget = `
SELECT * FROM expense
WHERE budget_id = $1
`

const upsertExpense = `
INSERT INTO expense (budget_id, category, amount, is_fixed, is_slack, id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (budget_id, category) DO UPDATE
SET amount = $3,
    is_fixed = $4,
	is_slack = $5
`

const deleteExpenseByBudget = `
DELETE FROM expense
WHERE budget_id = $1
`

func GetBudget(ctx context.Context, pool *pgxpool.Pool, owner, budgetID uuid.UUID) (*model.Budget, error) {
	var budgets []*model.Budget

	var err error

	if err = pgxscan.Select(ctx, pool, &budgets, getBudget, owner, budgetID); err == nil {
		if len(budgets) == 0 {
			err = errors.NoSuchElementError{Element: budgetID}
		} else {
			err = populateBudgets(ctx, pool, budgets)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch budget: %w", err)
	}

	return budgets[0], nil
}

func GetBudgets(ctx context.Context, pool *pgxpool.Pool, owner uuid.UUID, limit int) ([]*model.Budget, error) {
	var budgets []*model.Budget

	var err error

	if err = pgxscan.Select(ctx, pool, &budgets, getBudgetsByOwner, owner, limit); err == nil {
		err = populateBudgets(ctx, pool, budgets)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	return budgets, nil
}

func populateBudgets(ctx context.Context, pool *pgxpool.Pool, budgets []*model.Budget) error {
	// Batch budget loading
	batch := &pgx.Batch{}
	for _, b := range budgets {
		// get expenses
		batch.Queue(getExpensesForBudget, b.ID).Query(func(rows pgx.Rows) error {
			if err := pgxscan.ScanAll(&b.Expenses, rows); err != nil {
				return fmt.Errorf("failed to scan expenses: %w", err)
			}

			return nil
		})

		// get incomes
		batch.Queue(getIncomesByOwner, b.ID).Query(func(rows pgx.Rows) error {
			if err := pgxscan.ScanAll(&b.Incomes, rows); err != nil {
				return fmt.Errorf("failed to scan incomes: %w", err)
			}

			return nil
		})
	}

	if err := pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("get budgets batch failed: %w", err)
	}

	return nil
}

func PersistBudget(ctx context.Context, pool *pgxpool.Pool, budget *model.Budget) error {
	upsertBudgetQuery, upsertBudgetArgs, err := squirrel.Insert("budget").
		Values(budget.ID, ctxutil.GetUser(ctx), budget.Name).
		Suffix("ON CONFLICT (id, owner) DO UPDATE SET name = ?", budget.Name).
		ToSql()
	if err != nil {
		return fmt.Errorf("upsert budget SQL error: %w", err)
	}

	deleteIncomes, deleteIncomesArgs, err := squirrel.Delete("income").
		Where(squirrel.Eq{"owner": budget.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("delete incomes SQL error: %w", err)
	}

	deleteExpenses, deleteExpensesArgs, err := squirrel.Delete("expense").
		Where(squirrel.Eq{"budget_id": budget.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("delete expenses SQL error: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	if _, err = tx.Exec(ctx, upsertBudgetQuery, upsertBudgetArgs...); err != nil {
		return fmt.Errorf("failed to upsert budget: %w", err)
	}

	if _, err = tx.Exec(ctx, deleteIncomes, deleteIncomesArgs...); err != nil {
		return fmt.Errorf("failed to delete incomes: %w", err)
	}

	for _, income := range budget.Incomes {
		if _, err = tx.Exec(ctx, upsertIncome, income.Owner, income.Source, income.Amount); err != nil {
			return fmt.Errorf("failed to upsert income: %w", err)
		}
	}

	_, err = tx.Exec(ctx, deleteExpenses, deleteExpensesArgs...)
	if err != nil {
		return fmt.Errorf("failed to delete expenses: %w", err)
	}

	for _, expense := range budget.Expenses {
		expenseID := expense.ID
		if expenseID == uuid.Nil {
			expenseID = uuid.New()
		}

		_, err = tx.Exec(ctx, upsertExpense,
			expense.BudgetID, expense.Category, expense.Amount, expense.Fixed, expense.Slack, expenseID)
		if err != nil {
			return fmt.Errorf("failed to upsert expense: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction\n%w", err)
	}

	return nil
}

func DeleteBudget(ctx context.Context, pool *pgxpool.Pool, budget *model.Budget) error {
	batch := &pgx.Batch{}
	batch.Queue(deleteBudget, ctxutil.GetUser(ctx), budget.ID)
	batch.Queue(deleteIncomeByOwner, budget.ID)
	batch.Queue(deleteExpenseByBudget, budget.ID)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("batch operation failed during delete budget: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
