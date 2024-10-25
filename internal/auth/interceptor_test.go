package auth_test

import (
	"encoding/hex"
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

func TestInterceptorNoSID(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	handler := auth.Interceptor{}

	handler.ServeHTTP(w, &http.Request{})
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInterceptorSIDNotHex(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	handler := auth.Interceptor{}

	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost", nil)
	request.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: "(^_^) [o_o] (^.^) ($.$)",
	})
	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInterceptorInvalidSIDFormat(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	handler := auth.Interceptor{}

	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost", nil)
	request.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: "1a2b3c",
	})
	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInterceptorInvalidSID(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	w := httptest.NewRecorder()
	handler := auth.Interceptor{
		Pool: pool,
	}

	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost", nil)
	request.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: "1a2b3c1a2b3c1a2b1a2b3c1a2b3c1a2b",
	})
	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInterceptorValidSID(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	token := auth.NewSessionToken(user, time.Hour)

	// Save the token
	err := auth.SaveSessionToken(context.Background(), pool, token)
	require.NoError(t, err)

	// Make a handler that will write the user ID and SID to the recorder
	w := httptest.NewRecorder()
	handler := auth.Interceptor{
		Pool: pool,
		Intercepted: helper.FuncHandler{HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			u, _ := r.Context().Value(ctxutil.CTXUser).(uuid.UUID)
			sid, _ := r.Context().Value(ctxutil.CTXSID).(uuid.UUID)

			_, _ = w.Write([]byte(u.String() + sid.String()))
		}},
	}

	tokenIDBytes, err := token.ID.MarshalBinary()
	require.NoError(t, err)

	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost", nil)
	request.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: hex.EncodeToString(tokenIDBytes),
	})
	handler.ServeHTTP(w, request)
	require.Equal(t, user.String()+token.ID.String(), w.Body.String())
}
