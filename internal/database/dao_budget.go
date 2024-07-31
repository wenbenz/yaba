package database

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"yaba/internal/budget"
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

	var budgets []*budget.Budget
	if err := pgxscan.Select(ctx, pool, &budgets, getBudgetsByID, args...); err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	for _, b := range budgets {
		// get expenses
		var expenses []*budget.Expense
		if err := pgxscan.Select(ctx, pool, &expenses, getExpensesForBudget, b.ID); err != nil {
			return nil, fmt.Errorf("failed to get expenses for budget %s: %w", b.ID.String(), err)
		}

		b.Expenses = make(map[string]*budget.Expense)
		for _, e := range expenses {
			b.Expenses[e.Category] = e
		}

		// get incomes
		var incomes []*budget.Income
		if err := pgxscan.Select(ctx, pool, &incomes, getIncomesByOwner, b.ID); err != nil {
			return nil, fmt.Errorf("failed to get incomes for budget %s\n%w", b.ID.String(), err)
		}

		b.Incomes = make(map[string]*budget.Income)
		for _, income := range incomes {
			b.Incomes[income.Source] = income
		}
	}

	return budgets, nil
}

func PersistBudget(ctx context.Context, pool *pgxpool.Pool, budget *budget.Budget) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	if _, err = tx.Exec(ctx, upsertBudget, budget.ID, budget.Name, budget.Strategy); err != nil {
		return fmt.Errorf("could not upsert budget: %w", err)
	}

	if _, err = tx.Exec(ctx, deleteIncomeByOwner, budget.ID); err != nil {
		return fmt.Errorf("could not delete budget incomes: %w", err)
	}

	for _, income := range budget.Incomes {
		if _, err = tx.Exec(ctx, upsertIncome, income.Owner, income.Source, income.Amount); err != nil {
			return fmt.Errorf("could not save budget incomes: %w", err)
		}
	}

	if _, err = tx.Exec(ctx, deleteExpenseByBudget, budget.ID); err != nil {
		return fmt.Errorf("could not delete expenses: %w", err)
	}

	for _, expense := range budget.Expenses {
		if _, err = tx.Exec(ctx, upsertExpense,
			expense.BudgetID, expense.Category, expense.Amount, expense.Fixed, expense.Slack); err != nil {
			return fmt.Errorf("could not upsert expenses: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction\n%w", err)
	}

	return nil
}

func DeleteBudget(ctx context.Context, pool *pgxpool.Pool, budget *budget.Budget) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	if _, err = tx.Exec(ctx, deleteBudget, budget.ID); err != nil {
		return fmt.Errorf("could not delete budget: %w", err)
	}

	if _, err = tx.Exec(ctx, deleteIncomeByOwner, budget.ID); err != nil {
		return fmt.Errorf("could not delete budget incomes: %w", err)
	}

	if _, err = tx.Exec(ctx, deleteExpenseByBudget, budget.ID); err != nil {
		return fmt.Errorf("could not delete budget expenses: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
