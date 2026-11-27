package user

import (
	"bytes"
	"fmt"
	"yaba/errors"
	"yaba/internal/database"
	"yaba/internal/model"

	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

func CreateNewUser(
	ctx context.Context,
	pool *pgxpool.Pool,
	username string,
	password string,
) (*uuid.UUID, error) {
	var (
		passwordHash []byte
		err          error
	)

	if username == "" || password == "" {
		return nil, errors.InvalidInputError{Input: "username/password cannot be empty"}
	}

	id := uuid.New()

	if passwordHash, err = hashPassword(id, password); err == nil {
		err = database.CreateUser(ctx, pool, &model.User{
			ID:           id,
			Username:     username,
			PasswordHash: passwordHash,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &id, nil
}

func VerifyUser(
	ctx context.Context,
	pool *pgxpool.Pool,
	username, password string,
) (*uuid.UUID, error) {
	var (
		u   *model.User
		err error
	)

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

func ChangePassword(
	ctx context.Context,
	pool *pgxpool.Pool,
	id uuid.UUID,
	current, next string,
) error {
	u, err := database.GetUserByID(ctx, pool, id)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	hash, err := hashPassword(id, current)
	if err != nil {
		return err
	}

	if !bytes.Equal(hash, u.PasswordHash) {
		return errors.InvalidInputError{Input: "current password"}
	}

	if u.PasswordHash, err = hashPassword(id, next); err != nil {
		return err
	}

	if err = database.UpdateUser(ctx, pool, u); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

const passwordHashMemory = 64 * 1024
const passwordHashThreads = 4
const passwordHashKeylength = 32

func hashPassword(userID uuid.UUID, password string) ([]byte, error) {
	idBytes, err := userID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user id: %w", err)
	}

	return argon2.IDKey(
		[]byte(password),
		idBytes,
		1,
		passwordHashMemory,
		passwordHashThreads,
		passwordHashKeylength,
	), nil
}
