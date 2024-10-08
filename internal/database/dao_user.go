package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"yaba/internal/model"
)

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *model.User) error {
	sql, args, err := squirrel.
		Insert("user_profile").
		Values(user.ID, user.Username, user.PasswordHash).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err == nil {
		_, err = pool.Exec(ctx, sql, args...)
	}

	return err
}

func GetUserByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*model.User, error) {
	u := &model.User{}

	query, args, err := squirrel.
		Select("*").
		From("user_profile").
		Where(squirrel.Eq{"username": username}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err == nil {
		err = pgxscan.Get(ctx, pool, u, query, args...)
	}

	if u.ID == uuid.Nil {
		u = nil
	}

	return u, err
}
