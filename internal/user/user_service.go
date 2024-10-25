package user

import (
	"bytes"
	"fmt"
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

	if err != nil {
		return nil, err
	}

	err = database.CreateUser(ctx, pool, &model.User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
	})

	return &id, fmt.Errorf("failed to create user: %w", err)
}

func VerifyUser(ctx context.Context, pool *pgxpool.Pool, username, password string) (*uuid.UUID, error) {
	var u *model.User
	var err error

	if u, err = database.GetUserByUsername(ctx, pool, username); err == nil {
		var hash []byte
		if hash, err = hashPassword(u.ID, password); err == nil {
			if bytes.Equal(hash, u.PasswordHash) {
				return &u.ID, nil
			}
		}
	}

	return nil, err
}

const passwordHashMemory = 64 * 1024
const passwordHashThreads = 4
const passwordHashKeylength = 32

func hashPassword(userID uuid.UUID, password string) ([]byte, error) {
	idBytes, err := userID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user id: %w", err)
	}

	return argon2.IDKey([]byte(password), idBytes, 1, passwordHashMemory, passwordHashThreads, passwordHashKeylength), nil
}
