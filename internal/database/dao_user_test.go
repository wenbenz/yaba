package database_test

import (
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/test/helper"
)

func TestCreateUser(t *testing.T) {
	pool := helper.GetTestPool()
	user := &model.User{
		ID:           uuid.New(),
		Username:     gofakeit.Username(),
		PasswordHash: []byte(gofakeit.Password(true, true, true, true, true, 8)),
	}
	database.CreateUser(context.Background(), pool, user)

	fetched, err := database.GetUserByUsername(context.Background(), pool, user.Username)
	require.NoError(t, err)
	require.Equal(t, user.ID, fetched.ID)
	require.Equal(t, user.Username, fetched.Username)
	require.Equal(t, user.PasswordHash, fetched.PasswordHash)
}
