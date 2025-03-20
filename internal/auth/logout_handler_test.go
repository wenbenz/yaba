package auth_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"yaba/internal/auth"
	"yaba/internal/ctxutil"
	"yaba/internal/test/helper"
)

func TestLogoutHandler(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)

	// Save the token
	err := auth.SaveSessionToken(context.Background(), pool, token)
	require.NoError(t, err)

	// Create a request with the user context
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(ctxutil.WithUser(req.Context(), user))
	cookie, _ := auth.BakeCookie(token, "host")
	req.AddCookie(cookie)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create the handler
	handler := auth.NewLogoutHandler(pool)

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	require.Equal(t, http.StatusOK, rr.Code)

	// Check that the session token is deleted
	token, err = auth.GetSessionToken(context.Background(), pool, token.ID)
	require.Nil(t, token)
	require.ErrorContains(t, err, "failed to retrieve session token")

	// Check that the session cookie is cleared
	cookie = rr.Result().Cookies()[0] // nolint:bodyclose

	require.Equal(t, "sid", cookie.Name)
	require.Equal(t, "", cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
}

func TestLogoutHandler_InvalidSessionCookie(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()

	// Create a request with the user context and an invalid session cookie
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(ctxutil.WithUser(req.Context(), user))
	invalidCookie := &http.Cookie{
		Name:  "sid",
		Value: "invalid",
	}
	req.AddCookie(invalidCookie)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create the handler
	handler := auth.NewLogoutHandler(pool)

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogoutHandler_ErrorDeletingSessionToken(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)

	// Create a request with the user context and a valid session cookie
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(ctxutil.WithUser(req.Context(), user))
	cookie, _ := auth.BakeCookie(token, "host")
	req.AddCookie(cookie)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create the handler
	handler := auth.NewLogoutHandler(pool)

	// Simulate an error by closing the pool
	pool.Close()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
