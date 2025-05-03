package auth_test

import (
	"encoding/hex"
	"net/http"
	"testing"
	"time"
	"yaba/internal/auth"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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
	err := auth.SaveSessionToken(t.Context(), pool, token)
	require.NoError(t, err)

	// Fetch and assert it's the same
	fetched, err := auth.GetSessionToken(t.Context(), pool, token.ID)
	require.NoError(t, err)
	require.Equal(t, token.User, fetched.User)
	require.Equal(t, token.Type, fetched.Type)
	require.Equal(t, token.ID, fetched.ID)
	require.Equal(t, token.Created.Unix(), fetched.Created.Unix())
	require.Equal(t, token.Expires.Unix(), fetched.Expires.Unix())

	// Delete the token
	require.NoError(t, auth.DeleteSessionToken(t.Context(), pool, token.ID))

	// Fetch and assert it doesn't exist.
	fetched, err = auth.GetSessionToken(t.Context(), pool, token.ID)
	require.ErrorContains(t, err, "failed to retrieve session token")
	require.Nil(t, fetched)
}

func TestBakeCookie(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)
	cookie, err := auth.BakeCookie(token, "domain.com")

	require.NoError(t, err)
	require.Equal(t, "domain.com", cookie.Domain)
	require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	require.True(t, cookie.Secure)
	require.True(t, cookie.HttpOnly)
	require.Equal(t, token.Expires, cookie.Expires)

	decodedCookie, err := hex.DecodeString(cookie.Value)
	require.NoError(t, err)

	require.Equal(t, token.ID, decodedCookie)
}
