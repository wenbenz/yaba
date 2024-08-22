package database

import (
	"context"
	"fmt"
	"yaba/errors"
	"yaba/internal/budget"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getBudget = `
SELECT * FROM budget
WHERE id = $1;
`

const getBudgetsByOwner = `
SELECT * FROM budget
WHERE owner = $1
LIMIT $2;
`

const upsertBudget = `
INSERT INTO budget (id, owner, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET name = $3;
`

const deleteBudget = `
DELETE FROM budget
WHERE id = $1
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
INSERT INTO expense (budget_id, category, amount, is_fixed, is_slack)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (budget_id, category) DO UPDATE
SET amount = $3,
    is_fixed = $4,
	is_slack = $5
`

const deleteExpenseByBudget = `
DELETE FROM expense
WHERE budget_id = $1
`

func GetBudget(ctx context.Context, pool *pgxpool.Pool, budgetID uuid.UUID) (*budget.Budget, error) {
	var budgets []*budget.Budget

	var err error

	if err = pgxscan.Select(ctx, pool, &budgets, getBudget, budgetID); err == nil {
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

func GetBudgets(ctx context.Context, pool *pgxpool.Pool, owner uuid.UUID, limit int) ([]*budget.Budget, error) {
	var budgets []*budget.Budget

	var err error

	if err = pgxscan.Select(ctx, pool, &budgets, getBudgetsByOwner, owner, limit); err == nil {
		err = populateBudgets(ctx, pool, budgets)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	return budgets, nil
}

func populateBudgets(ctx context.Context, pool *pgxpool.Pool, budgets []*budget.Budget) error {
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

func PersistBudget(ctx context.Context, pool *pgxpool.Pool, budget *budget.Budget) error {
	batch := &pgx.Batch{}
	batch.Queue(upsertBudget, budget.ID, budget.Owner, budget.Name)
	batch.Queue(deleteIncomeByOwner, budget.ID)
	batch.Queue(deleteExpenseByBudget, budget.ID)

	for _, income := range budget.Incomes {
		batch.Queue(upsertIncome, income.Owner, income.Source, income.Amount)
	}

	for _, expense := range budget.Expenses {
		batch.Queue(upsertExpense, expense.BudgetID, expense.Category, expense.Amount, expense.Fixed, expense.Slack)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("batch operation failed while persisting budget: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction\n%w", err)
	}

	return nil
}

func DeleteBudget(ctx context.Context, pool *pgxpool.Pool, budget *budget.Budget) error {
	batch := &pgx.Batch{}
	batch.Queue(deleteBudget, budget.ID)
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
