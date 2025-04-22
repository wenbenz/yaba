package auth_test

import (
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"yaba/internal/auth"
	"yaba/internal/ctxutil"
	"yaba/internal/test/helper"
)

func TestInvalidSIDFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "empty",
			value: "",
		},
		{
			name:  "not hex",
			value: "(^_^) [o_o] (^.^) ($.$)",
		},
		{
			name:  "wrong length",
			value: "1a2b3c",
		},
	}

	for _, test := range tests {
		sid := test.value
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			handler := auth.SessionInterceptor{
				Intercepted: helper.FuncHandler{HandlerFunc: func(writer http.ResponseWriter, request *http.Request) {
					_, _ = writer.Write([]byte(ctxutil.GetUser(request.Context()).String()))
				}},
			}

			request, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "localhost", nil)
			request.AddCookie(&http.Cookie{
				Name:  "sid",
				Value: sid,
			})
			handler.ServeHTTP(w, request)
			require.Equal(t, "00000000-0000-0000-0000-000000000000", w.Body.String())
		})
	}
}

func TestInterceptorInvalidSID(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()

	w := httptest.NewRecorder()
	handler := auth.SessionInterceptor{
		Pool:        pool,
		Intercepted: auth.NewAuthRequired(helper.FuncHandler{}),
	}

	request, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "localhost", nil)
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
	err := auth.SaveSessionToken(t.Context(), pool, token)
	require.NoError(t, err)

	// Make a handler that will write the user ID and SID to the recorder
	w := httptest.NewRecorder()
	handler := auth.SessionInterceptor{
		Pool: pool,
		Intercepted: helper.FuncHandler{HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			u, _ := r.Context().Value(ctxutil.CTXUser).(uuid.UUID)
			sid, _ := r.Context().Value(ctxutil.CTXSID).([]byte)

			_, _ = w.Write([]byte(u.String() + hex.EncodeToString(sid)))
		}},
	}

	require.NoError(t, err)

	request, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "localhost", nil)
	request.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: hex.EncodeToString(token.ID),
	})
	handler.ServeHTTP(w, request)
	require.Equal(t, user.String()+hex.EncodeToString(token.ID), w.Body.String())
}
