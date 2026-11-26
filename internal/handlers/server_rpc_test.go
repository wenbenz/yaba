package handlers_test

import (
	"encoding/hex"
	"net/http/httptest"
	"testing"
	"time"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/auth"
	"yaba/internal/handlers"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRPCUsesSessionCookie(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	h, err := handlers.BuildServerHandler(pool)
	require.NoError(t, err)

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	client := yabav1connect.NewPaymentMethodServiceClient(srv.Client(), srv.URL+"/rpc")

	_, err = client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))

	token := auth.NewSessionToken(uuid.New(), time.Hour)
	require.NoError(t, auth.SaveSessionToken(t.Context(), pool, token))

	req := connect.NewRequest(&yabav1.ListPaymentMethodsRequest{})
	req.Header().Set("Cookie", "sid="+hex.EncodeToString(token.ID))

	_, err = client.ListPaymentMethods(t.Context(), req)
	require.NoError(t, err)
}
