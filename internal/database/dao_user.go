package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
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
	var u *model.User

	query, args, err := squirrel.
		Select("*").
		From("user_profile").
		Where(squirrel.Eq{"username": username}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err == nil {
		var users []*model.User
		err = pgxscan.Select(ctx, pool, &users, query, args...)
		if len(users) > 0 {
			u = users[0]
		}
	}

	return u, err
}
