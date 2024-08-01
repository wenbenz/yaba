package database

import (
	"context"
	"fmt"
	"time"

	"yaba/internal/budget"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const listExpenditures = `
SELECT * FROM expenditure
WHERE owner = $1
  AND date >= $2
  AND date <= $3
ORDER BY date
LIMIT $4;
`

const insertExpenditure = `
INSERT INTO expenditure (owner, name, amount, date, method, budget_category, reward_category, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func ListExpenditures(ctx context.Context, pool *pgxpool.Pool, owner uuid.UUID, since, until time.Time, limit int,
) ([]*budget.Expenditure, error) {
	var expenditures []*budget.Expenditure
	if err := pgxscan.Select(ctx, pool, &expenditures, listExpenditures, owner, since, until, limit); err != nil {
		return nil, fmt.Errorf("failed to get expenditures: %w", err)
	}

	return expenditures, nil
}

func PersistExpenditures(ctx context.Context, pool *pgxpool.Pool, expenditures []*budget.Expenditure,
) error {
	//nolint:exhaustruct
	batch := &pgx.Batch{}
	for _, e := range expenditures {
		batch.Queue(insertExpenditure, e.Owner, e.Name, e.Amount, e.Date,
			e.Method, e.BudgetCategory, e.RewardCategory, e.Comment)
	}

	if err := pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to save batch of expenditures: %w", err)
	}

	return nil
}
