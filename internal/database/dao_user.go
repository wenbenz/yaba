package database

import (
	"bytes"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
	"golang.org/x/net/context"
	"yaba/internal/model"
)

func CreateUser(ctx context.Context, pool *pgxpool.Pool, username, password string) error {
	id := uuid.New()
	passwordHash, err := hashPassword(id, password)
	if err == nil {
		sql, args, err := squirrel.
			Insert("user").
			Values(id, username, passwordHash).ToSql()

		if err == nil {
			_, err = pool.Exec(ctx, sql, args...)
		}
	}

	return err
}

func VerifyUser(ctx context.Context, pool *pgxpool.Pool, username, password string) (bool, error) {
	u, err := GetUserByUsername(ctx, pool, username)
	if err != nil {
		return false, err
	}

	hash, err := hashPassword(u.ID, password)

	return bytes.Equal(hash, u.PasswordHash), err
}

func GetUserByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*model.User, error) {
	var u *model.User

	err := squirrel.
		Select("id").
		From("user").
		Where(squirrel.Eq{"username": username}).
		RunWith(PgxpoolSquirrelAddapter{Pool: pool}).
		QueryRowContext(ctx).
		Scan(u)

	return u, err
}

func hashPassword(userId uuid.UUID, password string) ([]byte, error) {
	idBytes, err := userId.MarshalBinary()
	var passwordHash []byte
	if err == nil {
		passwordHash = argon2.IDKey([]byte(password), idBytes, 1, 64*1024, 4, 32)
	}
	return passwordHash, err
}
