package handlers_test

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/auth"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/handlers"
	"yaba/internal/model"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newRPCTestServer serves the full application handler, so requests go through SessionInterceptor and the
// real /rpc mount in BuildServerHandler.
func newRPCTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	h, err := handlers.BuildServerHandler(helper.GetTestPool())
	require.NoError(t, err)

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	return srv
}

func listPaymentMethodsWithCookie(t *testing.T, srv *httptest.Server, cookie string) (
	*connect.Response[yabav1.ListPaymentMethodsResponse], error,
) {
	t.Helper()

	client := yabav1connect.NewPaymentMethodServiceClient(srv.Client(), srv.URL+"/rpc")

	req := connect.NewRequest(&yabav1.ListPaymentMethodsRequest{})

	if cookie != "" {
		req.Header().Set("Cookie", cookie)
	}

	return client.ListPaymentMethods(t.Context(), req) //nolint:wrapcheck // test passthrough of the client error
}

func TestRPCListPaymentMethods_ValidSessionCookieSucceeds(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	srv := newRPCTestServer(t)
	user := uuid.New()

	token := auth.NewSessionToken(user, time.Hour)
	require.NotNil(t, token)
	require.NoError(t, auth.SaveSessionToken(t.Context(), pool, token))

	method := &model.PaymentMethod{ID: uuid.New(), DisplayName: "Session Card"}
	require.NoError(t, database.CreatePaymentMethod(ctxutil.WithUser(t.Context(), user), pool, method))

	res, err := listPaymentMethodsWithCookie(t, srv, "sid="+hex.EncodeToString(token.ID))
	require.NoError(t, err)
	require.Len(t, res.Msg.GetPaymentMethods(), 1)
	require.Equal(t, method.ID.String(), res.Msg.GetPaymentMethods()[0].GetId())
	require.Equal(t, "Session Card", res.Msg.GetPaymentMethods()[0].GetDisplayName())
}

func TestRPCListPaymentMethods_MissingOrInvalidSessionCookieIsUnauthenticated(t *testing.T) {
	t.Parallel()

	srv := newRPCTestServer(t)

	// Well-formed (32 bytes, hex) but never saved in the token table.
	unknown := make([]byte, 32)
	_, err := rand.Read(unknown)
	require.NoError(t, err)

	tests := []struct {
		name   string
		cookie string
	}{
		{name: "no cookie", cookie: ""},
		{name: "malformed cookie", cookie: "sid=(^_^)-not-hex"},
		{name: "wrong length cookie", cookie: "sid=1a2b3c"},
		{name: "unknown session token", cookie: "sid=" + hex.EncodeToString(unknown)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := listPaymentMethodsWithCookie(t, srv, test.cookie)
			require.Error(t, err)
			require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
		})
	}
}

func TestRPCListPaymentMethods_UnauthenticatedIsHTTP401WithConnectJSONBody(t *testing.T) {
	t.Parallel()

	srv := newRPCTestServer(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost,
		srv.URL+"/rpc"+yabav1connect.PaymentMethodServiceListPaymentMethodsProcedure, strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := srv.Client().Do(req)
	require.NoError(t, err)

	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	// A Connect error body, not the SPA index.html fallback.
	require.Contains(t, res.Header.Get("Content-Type"), "application/json")
}

func TestRPCUnknownProcedureIsNotFoundNotSPAFallback(t *testing.T) {
	t.Parallel()

	srv := newRPCTestServer(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost,
		srv.URL+"/rpc/"+yabav1connect.PaymentMethodServiceName+"/NoSuchMethod", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := srv.Client().Do(req)
	require.NoError(t, err)

	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusNotFound, res.StatusCode)
	require.NotContains(t, res.Header.Get("Content-Type"), "text/html")
}

func TestGraphQLWithoutSessionIsStillUnauthorized(t *testing.T) {
	t.Parallel()

	srv := newRPCTestServer(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+"/graphql",
		strings.NewReader(`{"query":"{ paymentMethods { id } }"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := srv.Client().Do(req)
	require.NoError(t, err)

	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
