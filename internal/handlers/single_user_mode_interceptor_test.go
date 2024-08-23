package handlers_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"yaba/internal/ctxutil"
	"yaba/internal/handlers"
)

type funcHandler struct {
	handlerFunc func(http.ResponseWriter, *http.Request)
}

func (h funcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handlerFunc(w, r)
}

func TestSingleUserModeDisabled(t *testing.T) {
	t.Parallel()

	intercepted := funcHandler{
		handlerFunc: func(_ http.ResponseWriter, _ *http.Request) {},
	}

	_, err := handlers.InterceptSingleUserMode(intercepted)
	require.ErrorContains(t, err, "Single user mode interceptor called, but SINGLE_USER_MODE is not enabled!")
}

func TestSingleUserModeNoUser(t *testing.T) {
	t.Setenv("SINGLE_USER_MODE", "true")

	intercepted := funcHandler{
		handlerFunc: func(_ http.ResponseWriter, _ *http.Request) {},
	}

	_, err := handlers.InterceptSingleUserMode(intercepted)
	require.ErrorContains(t, err, "could not parse UUID from SINGLE_USER_UUID in single user mode")
}

func TestSingleUserModeInterceptor(t *testing.T) {
	user := uuid.New()
	w := httptest.NewRecorder()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	intercepted := funcHandler{
		handlerFunc: func(w http.ResponseWriter, r *http.Request) {
			u, _ := r.Context().Value(ctxutil.CTXUser).(uuid.UUID)
			_, _ = w.Write([]byte(u.String()))
		},
	}

	handler, err := handlers.InterceptSingleUserMode(intercepted)
	require.NoError(t, err)

	handler.ServeHTTP(w, &http.Request{})

	require.Equal(t, user.String(), w.Body.String())
}
