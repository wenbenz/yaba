package rpc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/ctxutil"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/rpc"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// newClient returns a client whose requests carry user, standing in for auth.SessionInterceptor.
func newClient(t *testing.T, user uuid.UUID) yabav1connect.PaymentMethodServiceClient {
	t.Helper()

	h := rpc.NewHandler(helper.GetTestPool())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(ctxutil.WithUser(r.Context(), user)))
	}))
	t.Cleanup(srv.Close)

	return yabav1connect.NewPaymentMethodServiceClient(srv.Client(), srv.URL+"/rpc")
}

func newRewardCard(t *testing.T) *model.RewardCard {
	t.Helper()

	card := &model.RewardCard{ID: uuid.New(), Name: "Sapphire", Region: "US", Issuer: "Chase", RewardType: "points"}
	require.NoError(t, database.CreateRewardCard(t.Context(), helper.GetTestPool(), card))

	return card
}

func create(
	t *testing.T,
	client yabav1connect.PaymentMethodServiceClient,
	req *yabav1.CreatePaymentMethodRequest,
) *yabav1.PaymentMethod {
	t.Helper()

	res, err := client.CreatePaymentMethod(t.Context(), connect.NewRequest(req))
	require.NoError(t, err)

	return res.Msg.GetPaymentMethod()
}

func list(t *testing.T, client yabav1connect.PaymentMethodServiceClient) []*yabav1.PaymentMethod {
	t.Helper()

	res, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)

	return res.Msg.GetPaymentMethods()
}

func TestCreateAndListPaymentMethods(t *testing.T) {
	t.Parallel()

	card := newRewardCard(t)
	client := newClient(t, uuid.New())
	acquired := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)

	require.Empty(t, list(t, client))

	named := create(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:   card.ID.String(),
		DisplayName:  new("My Card"),
		AcquiredDate: timestamppb.New(acquired),
	})
	unnamed := create(t, client, &yabav1.CreatePaymentMethodRequest{CardTypeId: card.ID.String()})

	require.Equal(t, "Sapphire", unnamed.GetDisplayName())

	methods := list(t, client)
	require.Len(t, methods, 2)
	require.Equal(t, named.GetId(), methods[0].GetId())
	require.Equal(t, "My Card", methods[0].GetDisplayName())
	require.Equal(t, card.ID.String(), methods[0].GetCardTypeId())
	require.Equal(t, acquired, methods[0].GetAcquiredDate().AsTime())
	require.Nil(t, methods[0].GetCancelByDate())
	require.Equal(t, unnamed.GetId(), methods[1].GetId())

	require.Empty(t, list(t, newClient(t, uuid.New())), "other users must not see these")
}

func TestCreatePaymentMethod_InvalidArgument(t *testing.T) {
	t.Parallel()

	client := newClient(t, uuid.New())

	for name, req := range map[string]*yabav1.CreatePaymentMethodRequest{
		"unknown card type": {CardTypeId: uuid.NewString()},
		"malformed id":      {CardTypeId: "not-a-uuid"},
		"empty name":        {CardTypeId: newRewardCard(t).ID.String(), DisplayName: new("")},
	} {
		_, err := client.CreatePaymentMethod(t.Context(), connect.NewRequest(req))
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), name)
	}
}

func TestUpdatePaymentMethod(t *testing.T) {
	t.Parallel()

	card := newRewardCard(t)
	client := newClient(t, uuid.New())
	method := create(t, client, &yabav1.CreatePaymentMethodRequest{CardTypeId: card.ID.String()})

	update := &yabav1.UpdatePaymentMethodRequest{
		Id:          method.GetId(),
		CardTypeId:  card.ID.String(),
		DisplayName: new("Renamed"),
	}

	_, err := newClient(t, uuid.New()).UpdatePaymentMethod(t.Context(), connect.NewRequest(update))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err), "other users can't update it")

	res, err := client.UpdatePaymentMethod(t.Context(), connect.NewRequest(update))
	require.NoError(t, err)
	require.Equal(t, "Renamed", res.Msg.GetPaymentMethod().GetDisplayName())
	require.Equal(t, "Renamed", list(t, client)[0].GetDisplayName())
}

func TestDeletePaymentMethod(t *testing.T) {
	t.Parallel()

	client := newClient(t, uuid.New())
	method := create(t, client, &yabav1.CreatePaymentMethodRequest{CardTypeId: newRewardCard(t).ID.String()})
	req := connect.NewRequest(&yabav1.DeletePaymentMethodRequest{Id: method.GetId()})

	_, err := newClient(t, uuid.New()).DeletePaymentMethod(t.Context(), req)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err), "other users can't delete it")

	_, err = client.DeletePaymentMethod(t.Context(), req)
	require.NoError(t, err)
	require.Empty(t, list(t, client))

	_, err = client.DeletePaymentMethod(t.Context(), req)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestPaymentMethodService_RequiresUser(t *testing.T) {
	t.Parallel()

	client := newClient(t, uuid.Nil)

	_, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))

	// Auth runs before validation.
	_, err = client.DeletePaymentMethod(t.Context(), connect.NewRequest(&yabav1.DeletePaymentMethodRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
