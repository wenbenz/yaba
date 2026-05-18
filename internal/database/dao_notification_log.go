package database

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WasNotificationSent(ctx context.Context, pool *pgxpool.Pool, pmID uuid.UUID, renewalYear int) (bool, error) {
	query, args, err := squirrel.
		Select("1").
		From("notification_log").
		Where(squirrel.Eq{"payment_method_id": pmID, "renewal_year": renewalYear}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to form query: %w", err)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to query notification log: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func LogNotificationSent(ctx context.Context, pool *pgxpool.Pool, pmID, userID uuid.UUID, renewalYear int) error {
	sql, args, err := squirrel.
		Insert("notification_log").
		Columns("payment_method_id", "user_id", "renewal_year").
		Values(pmID, userID, renewalYear).
		Suffix("ON CONFLICT DO NOTHING").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to construct sql: %w", err)
	}

	if _, err = pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to log notification: %w", err)
	}

	return nil
}
