package auth_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
	"time"
	"yaba/internal/auth"
	"yaba/internal/test/helper"
)

func TestNewToken(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	before := time.Now()
	token := auth.NewSessionToken(user, time.Hour)
	after := time.Now()

	require.Equal(t, user, token.User)
	require.Equal(t, "SESSION", token.Type)
	require.NotNil(t, token.ID)
	require.NotEqual(t, uuid.Nil, token.ID)

	require.LessOrEqual(t, before, token.Created)
	require.GreaterOrEqual(t, after, token.Created)

	require.LessOrEqual(t, before.Add(time.Hour), token.Expires)
	require.GreaterOrEqual(t, after.Add(time.Hour), token.Expires)
}

func TestTokenStorage(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)

	// Save the token
	err := auth.SaveSessionToken(context.Background(), pool, token)
	require.NoError(t, err)

	// Fetch and assert it's the same
	fetched, err := auth.GetSessionToken(context.Background(), pool, token.ID)
	require.NoError(t, err)
	require.Equal(t, token.User, fetched.User)
	require.Equal(t, token.Type, fetched.Type)
	require.Equal(t, token.ID, fetched.ID)
	require.Equal(t, token.Created.Unix(), fetched.Created.Unix())
	require.Equal(t, token.Expires.Unix(), fetched.Expires.Unix())

	// Delete the token
	require.NoError(t, auth.DeleteSessionToken(context.Background(), pool, token.ID))

	// Fetch and assert it doesn't exist.
	fetched, err = auth.GetSessionToken(context.Background(), pool, token.ID)
	require.ErrorContains(t, err, "failed to retrieve session token")
	require.Nil(t, fetched)
}
