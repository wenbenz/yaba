package user

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
	"golang.org/x/net/context"
	"yaba/internal/database"
	"yaba/internal/model"
)

func CreateNewUser(ctx context.Context, pool *pgxpool.Pool, username string, password string) (*uuid.UUID, error) {
	id := uuid.New()
	passwordHash, err := hashPassword(id, password)
	if err == nil {
		err = database.CreateUser(ctx, pool, &model.User{
			ID:           id,
			Username:     username,
			PasswordHash: passwordHash,
		})
	}

	return &id, err
}

func VerifyUser(ctx context.Context, pool *pgxpool.Pool, username, password string) (bool, error) {
	u, err := database.GetUserByUsername(ctx, pool, username)
	if err != nil {
		return false, err
	}

	hash, err := hashPassword(u.ID, password)

	return bytes.Equal(hash, u.PasswordHash), err
}

func hashPassword(userId uuid.UUID, password string) ([]byte, error) {
	idBytes, err := userId.MarshalBinary()
	var passwordHash []byte
	if err == nil {
		passwordHash = argon2.IDKey([]byte(password), idBytes, 1, 64*1024, 4, 32)
	}
	return passwordHash, err
}
