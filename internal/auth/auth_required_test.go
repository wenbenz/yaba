package auth_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
	"yaba/internal/auth"
	"yaba/internal/ctxutil"
	"yaba/internal/test/helper"
)

func TestAuthRequired(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	var handler http.Handler = helper.FuncHandler{HandlerFunc: func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}}

	request, err := http.NewRequestWithContext(context.WithValue(context.Background(), ctxutil.CTXUser, uuid.New()),
		http.MethodGet, "/", nil)
	require.NoError(t, err)

	handler = auth.NewAuthRequired(handler)
	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequiredNoUserInContext(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	var handler http.Handler = helper.FuncHandler{HandlerFunc: func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	handler = auth.NewAuthRequired(handler)
	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
