package database

import (
	"context"
	"fmt"

	"yaba/internal/budget"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getBudgetsByID = `
SELECT * FROM budget
WHERE id IN ($1)
`

const upsertBudget = `
INSERT INTO budget (id, name, strategy)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET name = $2,
    strategy = $3;
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

func GetBudgets(ctx context.Context, pool *pgxpool.Pool, ids []uuid.UUID) ([]*budget.Budget, error) {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	// First, see which budgets actually exist
	var budgets []*budget.Budget
	if err := pgxscan.Select(ctx, pool, &budgets, getBudgetsByID, args...); err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	// Batch budget loading
	batch := &pgx.Batch{}
	for _, b := range budgets {
		// get expenses
		batch.Queue(getExpensesForBudget, b.ID).Query(func(rows pgx.Rows) error {
			var expenses []*budget.Expense
			if err := pgxscan.ScanAll(&expenses, rows); err != nil {
				return fmt.Errorf("failed to scan expenses: %w", err)
			}

			b.Expenses = make(map[string]*budget.Expense)
			for _, e := range expenses {
				b.Expenses[e.Category] = e
			}

			return nil
		})

		// get incomes
		batch.Queue(getIncomesByOwner, b.ID).Query(func(rows pgx.Rows) error {
			var incomes []*budget.Income
			if err := pgxscan.ScanAll(&incomes, rows); err != nil {
				return fmt.Errorf("failed to scan incomes: %w", err)
			}

			b.Incomes = make(map[string]*budget.Income)
			for _, income := range incomes {
				b.Incomes[income.Source] = income
			}

			return nil
		})
	}

	if err := pool.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("get budgets batch failed: %w", err)
	}

	return budgets, nil
}

func PersistBudget(ctx context.Context, pool *pgxpool.Pool, budget *budget.Budget) error {
	batch := &pgx.Batch{}
	batch.Queue(upsertBudget, budget.ID, budget.Name, budget.Strategy)
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
