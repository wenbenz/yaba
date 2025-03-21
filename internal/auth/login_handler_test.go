package auth_test

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"yaba/internal/auth"
	"yaba/internal/test/helper"
	"yaba/internal/user"
)

func TestLoginHandler(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	testCases := []struct {
		name      string
		handlerFn func(*pgxpool.Pool) *auth.LoginHandler
		username  string
		password  string

		setupFn  func()
		assertFn func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "new user",
			handlerFn: auth.NewUserHandler,
			username:  "foo",
			password:  "bar",
			assertFn:  assertSuccess,
		},
		{
			name:      "new user no password",
			handlerFn: auth.NewUserHandler,
			username:  "foo",
			assertFn:  assertFail,
		},
		{
			name:      "login as existing user",
			handlerFn: auth.VerifyUserHandler,
			username:  "username-login",
			password:  "bar",

			setupFn: func() {
				_, _ = user.CreateNewUser(t.Context(), pool, "username-login", "bar")
			},
			assertFn: assertSuccess,
		},
		{
			name:      "login as non-existing user",
			handlerFn: auth.VerifyUserHandler,
			username:  "username-login",
			password:  "bar",
			assertFn:  assertFail,
		},
		{
			name:      "login as existing user no password",
			handlerFn: auth.VerifyUserHandler,
			username:  "username-login",

			setupFn: func() {
				_, _ = user.CreateNewUser(t.Context(), pool, "username-login", "bar")
			},
			assertFn: assertFail,
		},
		{
			name:      "login as existing user wrong password",
			handlerFn: auth.VerifyUserHandler,
			username:  "username-login",
			password:  "baz",

			setupFn: func() {
				_, _ = user.CreateNewUser(t.Context(), pool, "username-login", "bar")
			},
			assertFn: assertFail,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.setupFn != nil {
				tc.setupFn()
			}

			w := httptest.NewRecorder()
			handler := tc.handlerFn(pool)

			form := url.Values{}
			form.Add("username", tc.username)
			form.Add("password", tc.password)
			request, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "localhost",
				strings.NewReader(form.Encode()))
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			handler.ServeHTTP(w, request)
			tc.assertFn(t, w)
		})
	}
}

func assertSuccess(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	require.Equal(t, http.StatusFound, w.Code)
	require.NotEmpty(t, w.Header().Get("Set-Cookie"))
}

func assertFail(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Empty(t, w.Header().Get("Set-Cookie"))
}
