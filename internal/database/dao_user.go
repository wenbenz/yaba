package database

import (
	"fmt"
	"yaba/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *model.User) error {
	sql, args, err := squirrel.
		Insert("user_profile").
		Values(user.ID, user.Username, user.PasswordHash).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to construct sql: %w", err)
	}

	if _, err = pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func GetUserByUsername(
	ctx context.Context,
	pool *pgxpool.Pool,
	username string,
) (*model.User, error) {
	u := &model.User{}

	query, args, err := squirrel.
		Select("*").
		From("user_profile").
		Where(squirrel.Eq{"username": username}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to form query: %w", err)
	}

	if err = pgxscan.Get(ctx, pool, u, query, args...); err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if u.ID == uuid.Nil {
		u = nil
	}

	return u, nil
}
