package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"yaba/internal/auth"
	"yaba/internal/ctxutil"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLogoutHandler(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)

	// Save the token
	err := auth.SaveSessionToken(t.Context(), pool, token)
	require.NoError(t, err)

	// Create a request with the user context
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "/logout", nil)
	req = req.WithContext(ctxutil.WithUser(req.Context(), user))
	req.Host = "yaba.example.com"
	cookie, _ := auth.BakeCookie(token)
	req.AddCookie(cookie)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create the handler
	handler := auth.NewLogoutHandler(pool)

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/", rr.Header().Get("Location"))

	// Check that the session token is deleted
	token, err = auth.GetSessionToken(t.Context(), pool, token.ID)
	require.Nil(t, token)
	require.ErrorContains(t, err, "failed to retrieve session token")

	// Check that both the host-only and the legacy Domain session cookies are cleared
	cookies := rr.Result().Cookies() // nolint:bodyclose
	require.Len(t, cookies, 2)

	for _, c := range cookies {
		require.Equal(t, "sid", c.Name)
		require.Empty(t, c.Value)
		require.Equal(t, -1, c.MaxAge)
	}

	require.Empty(t, cookies[0].Domain)
	require.Equal(t, "yaba.example.com", cookies[1].Domain)
}

func TestLogoutHandlerNoCookie(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()

	auth.NewLogoutHandler(helper.GetTestPool()).ServeHTTP(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/", rr.Header().Get("Location"))
}
