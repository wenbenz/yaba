package database_test

import (
	"testing"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/test/helper"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := &model.User{
		ID:           uuid.New(),
		Username:     gofakeit.Username(),
		PasswordHash: []byte(gofakeit.Password(true, true, true, true, true, 8)),
	}

	err := database.CreateUser(t.Context(), pool, user)
	require.NoError(t, err)

	fetched, err := database.GetUserByUsername(t.Context(), pool, user.Username)
	require.NoError(t, err)
	require.Equal(t, user.ID, fetched.ID)
	require.Equal(t, user.Username, fetched.Username)
	require.Equal(t, user.PasswordHash, fetched.PasswordHash)
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := &model.User{
		ID:           uuid.New(),
		Username:     gofakeit.Username(),
		PasswordHash: []byte(gofakeit.Password(true, true, true, true, true, 8)),
	}

	require.NoError(t, database.CreateUser(t.Context(), pool, user))

	fetched, err := database.GetUserByID(t.Context(), pool, user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, user.ID, fetched.ID)
	require.Equal(t, user.Username, fetched.Username)
}

func TestUpdateUserProfile(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := &model.User{
		ID:           uuid.New(),
		Username:     gofakeit.Username(),
		PasswordHash: []byte(gofakeit.Password(true, true, true, true, true, 8)),
	}

	require.NoError(t, database.CreateUser(t.Context(), pool, user))

	email := gofakeit.Email()
	updated, err := database.UpdateUserProfile(t.Context(), pool, user.ID, &email, true)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, &email, updated.Email)
	require.True(t, updated.EmailRemindersEnabled)

	updated2, err := database.UpdateUserProfile(t.Context(), pool, user.ID, nil, false)
	require.NoError(t, err)
	require.Nil(t, updated2.Email)
	require.False(t, updated2.EmailRemindersEnabled)
}
