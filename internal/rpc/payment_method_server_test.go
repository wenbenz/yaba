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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// newTestClient serves the payment method service under /rpc on an httptest server. When user is not
// uuid.Nil, a middleware puts it in every request context, standing in for auth.SessionInterceptor.
func newTestClient(t *testing.T, pool *pgxpool.Pool, user uuid.UUID) yabav1connect.PaymentMethodServiceClient {
	t.Helper()

	path, handler := rpc.NewHandler(pool)

	mux := http.NewServeMux()
	mux.Handle("/rpc"+path, http.StripPrefix("/rpc", handler))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user != uuid.Nil {
			r = r.WithContext(ctxutil.WithUser(r.Context(), user))
		}

		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	return yabav1connect.NewPaymentMethodServiceClient(srv.Client(), srv.URL+"/rpc")
}

func newTestRewardCard(t *testing.T, pool *pgxpool.Pool, name string) uuid.UUID {
	t.Helper()

	card := &model.RewardCard{
		ID:         uuid.New(),
		Name:       name,
		Region:     "US",
		Issuer:     "Chase",
		RewardType: "points",
	}
	require.NoError(t, database.CreateRewardCard(t.Context(), pool, card))

	return card.ID
}

func createTestPaymentMethod(
	t *testing.T,
	client yabav1connect.PaymentMethodServiceClient,
	req *yabav1.CreatePaymentMethodRequest,
) *yabav1.PaymentMethod {
	t.Helper()

	res, err := client.CreatePaymentMethod(t.Context(), connect.NewRequest(req))
	require.NoError(t, err)

	pm := res.Msg.GetPaymentMethod()
	require.NotNil(t, pm)

	return pm
}

func requireTimestamp(t *testing.T, want time.Time, got *timestamppb.Timestamp) {
	t.Helper()

	require.NotNil(t, got)
	require.Equal(t, want.UTC(), got.AsTime())
}

func TestListPaymentMethods_EmptyListIsSuccess(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, helper.GetTestPool(), uuid.New())

	res, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Empty(t, res.Msg.GetPaymentMethods())
}

func TestListPaymentMethods_ReturnsOnlyCallersMethodsOrderedByDisplayName(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	client := newTestClient(t, pool, uuid.New())
	otherClient := newTestClient(t, pool, uuid.New())

	acquiredB := time.Date(2024, 3, 21, 0, 0, 0, 0, time.UTC)
	cancelB := time.Date(2025, 3, 21, 0, 0, 0, 0, time.UTC)

	// Created out of order so that the ordering assertion is meaningful.
	cardB := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:   cardID.String(),
		DisplayName:  proto.String("Card B"),
		AcquiredDate: timestamppb.New(acquiredB),
		CancelByDate: timestamppb.New(cancelB),
	})
	cardA := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Card A"),
	})
	otherCard := createTestPaymentMethod(t, otherClient, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Another User's Card"),
	})

	res, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)

	methods := res.Msg.GetPaymentMethods()
	require.Len(t, methods, 2)
	require.Equal(t, cardA.GetId(), methods[0].GetId())
	require.Equal(t, "Card A", methods[0].GetDisplayName())
	require.Equal(t, cardID.String(), methods[0].GetCardTypeId())
	require.Nil(t, methods[0].GetAcquiredDate())
	require.Nil(t, methods[0].GetCancelByDate())
	require.Equal(t, cardB.GetId(), methods[1].GetId())
	require.Equal(t, "Card B", methods[1].GetDisplayName())
	require.Equal(t, cardID.String(), methods[1].GetCardTypeId())
	requireTimestamp(t, acquiredB, methods[1].GetAcquiredDate())
	requireTimestamp(t, cancelB, methods[1].GetCancelByDate())

	otherRes, err := otherClient.ListPaymentMethods(t.Context(),
		connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Len(t, otherRes.Msg.GetPaymentMethods(), 1)
	require.Equal(t, otherCard.GetId(), otherRes.Msg.GetPaymentMethods()[0].GetId())
}

func TestListPaymentMethods_LegacyRowWithNullCardTypeHasEmptyCardTypeID(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	client := newTestClient(t, pool, user)

	// Rows migrated by 000011 only have id, owner and display_name; card_type and the dates are NULL.
	legacyID := uuid.New()
	_, err := pool.Exec(t.Context(),
		"INSERT INTO payment_method (id, owner, display_name) VALUES ($1, $2, $3)",
		legacyID, user, "Legacy Card")
	require.NoError(t, err)

	res, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)

	methods := res.Msg.GetPaymentMethods()
	require.Len(t, methods, 1)
	require.Equal(t, legacyID.String(), methods[0].GetId())
	require.Equal(t, "Legacy Card", methods[0].GetDisplayName())
	require.Empty(t, methods[0].GetCardTypeId())
	require.Nil(t, methods[0].GetAcquiredDate())
	require.Nil(t, methods[0].GetCancelByDate())
}

func TestCreatePaymentMethod_SuccessReturnsAndPersistsRecord(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	client := newTestClient(t, pool, user)

	acquired := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	cancelBy := time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

	pm := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:   cardID.String(),
		DisplayName:  proto.String("Test Credit Card"),
		AcquiredDate: timestamppb.New(acquired),
		CancelByDate: timestamppb.New(cancelBy),
	})

	id, err := uuid.Parse(pm.GetId())
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
	require.Equal(t, "Test Credit Card", pm.GetDisplayName())
	require.Equal(t, cardID.String(), pm.GetCardTypeId())
	requireTimestamp(t, acquired, pm.GetAcquiredDate())
	requireTimestamp(t, cancelBy, pm.GetCancelByDate())

	persisted, err := database.GetPaymentMethod(ctxutil.WithUser(t.Context(), user), pool, id)
	require.NoError(t, err)
	require.Equal(t, user, persisted.Owner)
	require.Equal(t, "Test Credit Card", persisted.DisplayName)
	require.Equal(t, cardID, persisted.CardType)
	require.True(t, persisted.AcquiredDate.Valid)
	require.Equal(t, acquired, persisted.AcquiredDate.Time.UTC())
	require.True(t, persisted.CancelByDate.Valid)
	require.Equal(t, cancelBy, persisted.CancelByDate.Time.UTC())
}

func TestCreatePaymentMethod_MissingDisplayNameDefaultsToRewardCardName(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	cardID := newTestRewardCard(t, pool, "Amex Gold")
	client := newTestClient(t, pool, user)

	pm := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId: cardID.String(),
	})
	require.Equal(t, "Amex Gold", pm.GetDisplayName())
	require.Nil(t, pm.GetAcquiredDate())
	require.Nil(t, pm.GetCancelByDate())

	// Unset timestamps must be stored as NULL.
	persisted, err := database.GetPaymentMethod(ctxutil.WithUser(t.Context(), user), pool,
		uuid.MustParse(pm.GetId()))
	require.NoError(t, err)
	require.Equal(t, "Amex Gold", persisted.DisplayName)
	require.False(t, persisted.AcquiredDate.Valid)
	require.False(t, persisted.CancelByDate.Valid)
}

func TestCreatePaymentMethod_UnknownCardTypeIsInvalidArgument(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, helper.GetTestPool(), uuid.New())

	_, err := client.CreatePaymentMethod(t.Context(), connect.NewRequest(&yabav1.CreatePaymentMethodRequest{
		CardTypeId:  uuid.NewString(),
		DisplayName: proto.String("Ghost Card"),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

	res, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Empty(t, res.Msg.GetPaymentMethods())
}

func TestCreatePaymentMethod_InvalidRequestIsInvalidArgument(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Freedom")
	client := newTestClient(t, pool, uuid.New())

	tests := []struct {
		name string
		req  *yabav1.CreatePaymentMethodRequest
	}{
		{
			name: "card_type_id is not a uuid",
			req:  &yabav1.CreatePaymentMethodRequest{CardTypeId: "not-a-uuid"},
		},
		{
			name: "card_type_id is missing",
			req:  &yabav1.CreatePaymentMethodRequest{DisplayName: proto.String("No Card Type")},
		},
		{
			name: "display_name is present but empty",
			req:  &yabav1.CreatePaymentMethodRequest{CardTypeId: cardID.String(), DisplayName: proto.String("")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.CreatePaymentMethod(t.Context(), connect.NewRequest(test.req))
			require.Error(t, err)
			require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestUpdatePaymentMethod_SuccessReplacesRecord(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	originalCardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	newCardID := newTestRewardCard(t, pool, "Amex Gold")
	client := newTestClient(t, pool, user)

	created := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:   originalCardID.String(),
		DisplayName:  proto.String("Original Card"),
		AcquiredDate: timestamppb.New(time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)),
		CancelByDate: timestamppb.New(time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)),
	})

	acquired := time.Date(2024, 3, 21, 0, 0, 0, 0, time.UTC)
	cancelBy := time.Date(2025, 3, 21, 0, 0, 0, 0, time.UTC)

	res, err := client.UpdatePaymentMethod(t.Context(), connect.NewRequest(&yabav1.UpdatePaymentMethodRequest{
		Id:           created.GetId(),
		CardTypeId:   newCardID.String(),
		DisplayName:  proto.String("Updated Card"),
		AcquiredDate: timestamppb.New(acquired),
		CancelByDate: timestamppb.New(cancelBy),
	}))
	require.NoError(t, err)

	updated := res.Msg.GetPaymentMethod()
	require.NotNil(t, updated)
	require.Equal(t, created.GetId(), updated.GetId())
	require.Equal(t, "Updated Card", updated.GetDisplayName())
	require.Equal(t, newCardID.String(), updated.GetCardTypeId())
	requireTimestamp(t, acquired, updated.GetAcquiredDate())
	requireTimestamp(t, cancelBy, updated.GetCancelByDate())

	persisted, err := database.GetPaymentMethod(ctxutil.WithUser(t.Context(), user), pool,
		uuid.MustParse(created.GetId()))
	require.NoError(t, err)
	require.Equal(t, "Updated Card", persisted.DisplayName)
	require.Equal(t, newCardID, persisted.CardType)
	require.Equal(t, acquired, persisted.AcquiredDate.Time.UTC())
	require.Equal(t, cancelBy, persisted.CancelByDate.Time.UTC())

	list, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Len(t, list.Msg.GetPaymentMethods(), 1)
	require.Equal(t, created.GetId(), list.Msg.GetPaymentMethods()[0].GetId())
	require.Equal(t, "Updated Card", list.Msg.GetPaymentMethods()[0].GetDisplayName())
}

func TestUpdatePaymentMethod_OtherUsersIDIsNotFoundAndLeavesRecordUnchanged(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	owner := uuid.New()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	ownerClient := newTestClient(t, pool, owner)
	otherClient := newTestClient(t, pool, uuid.New())

	created := createTestPaymentMethod(t, ownerClient, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Owner's Card"),
	})

	_, err := otherClient.UpdatePaymentMethod(t.Context(), connect.NewRequest(&yabav1.UpdatePaymentMethodRequest{
		Id:          created.GetId(),
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Hijacked"),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

	persisted, err := database.GetPaymentMethod(ctxutil.WithUser(t.Context(), owner), pool,
		uuid.MustParse(created.GetId()))
	require.NoError(t, err)
	require.Equal(t, "Owner's Card", persisted.DisplayName)
}

func TestUpdatePaymentMethod_MissingIDIsNotFound(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	client := newTestClient(t, pool, uuid.New())

	_, err := client.UpdatePaymentMethod(t.Context(), connect.NewRequest(&yabav1.UpdatePaymentMethodRequest{
		Id:          uuid.NewString(),
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Does Not Exist"),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestUpdatePaymentMethod_UnknownCardTypeIsInvalidArgument(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	user := uuid.New()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	client := newTestClient(t, pool, user)

	created := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Original Card"),
	})

	_, err := client.UpdatePaymentMethod(t.Context(), connect.NewRequest(&yabav1.UpdatePaymentMethodRequest{
		Id:          created.GetId(),
		CardTypeId:  uuid.NewString(),
		DisplayName: proto.String("Updated Card"),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

	persisted, err := database.GetPaymentMethod(ctxutil.WithUser(t.Context(), user), pool,
		uuid.MustParse(created.GetId()))
	require.NoError(t, err)
	require.Equal(t, "Original Card", persisted.DisplayName)
	require.Equal(t, cardID, persisted.CardType)
}

func TestUpdatePaymentMethod_InvalidRequestIsInvalidArgument(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Freedom")
	client := newTestClient(t, pool, uuid.New())

	tests := []struct {
		name string
		req  *yabav1.UpdatePaymentMethodRequest
	}{
		{
			name: "id is not a uuid",
			req:  &yabav1.UpdatePaymentMethodRequest{Id: "not-a-uuid", CardTypeId: cardID.String()},
		},
		{
			name: "id is missing",
			req:  &yabav1.UpdatePaymentMethodRequest{CardTypeId: cardID.String()},
		},
		{
			name: "card_type_id is not a uuid",
			req:  &yabav1.UpdatePaymentMethodRequest{Id: uuid.NewString(), CardTypeId: "not-a-uuid"},
		},
		{
			name: "display_name is present but empty",
			req: &yabav1.UpdatePaymentMethodRequest{
				Id:          uuid.NewString(),
				CardTypeId:  cardID.String(),
				DisplayName: proto.String(""),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.UpdatePaymentMethod(t.Context(), connect.NewRequest(test.req))
			require.Error(t, err)
			require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestDeletePaymentMethod_SuccessThenSecondDeleteIsNotFound(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	client := newTestClient(t, pool, uuid.New())

	created := createTestPaymentMethod(t, client, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Test Credit Card"),
	})
	deleteReq := &yabav1.DeletePaymentMethodRequest{Id: created.GetId()}

	_, err := client.DeletePaymentMethod(t.Context(), connect.NewRequest(deleteReq))
	require.NoError(t, err)

	list, err := client.ListPaymentMethods(t.Context(), connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Empty(t, list.Msg.GetPaymentMethods())

	_, err = client.DeletePaymentMethod(t.Context(), connect.NewRequest(deleteReq))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestDeletePaymentMethod_OtherUsersIDIsNotFoundAndLeavesRecord(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	cardID := newTestRewardCard(t, pool, "Chase Sapphire Reserve")
	ownerClient := newTestClient(t, pool, uuid.New())
	otherClient := newTestClient(t, pool, uuid.New())

	created := createTestPaymentMethod(t, ownerClient, &yabav1.CreatePaymentMethodRequest{
		CardTypeId:  cardID.String(),
		DisplayName: proto.String("Owner's Card"),
	})

	_, err := otherClient.DeletePaymentMethod(t.Context(),
		connect.NewRequest(&yabav1.DeletePaymentMethodRequest{Id: created.GetId()}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

	list, err := ownerClient.ListPaymentMethods(t.Context(),
		connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))
	require.NoError(t, err)
	require.Len(t, list.Msg.GetPaymentMethods(), 1)
	require.Equal(t, created.GetId(), list.Msg.GetPaymentMethods()[0].GetId())
}

func TestDeletePaymentMethod_NonUUIDIDIsInvalidArgument(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, helper.GetTestPool(), uuid.New())

	_, err := client.DeletePaymentMethod(t.Context(),
		connect.NewRequest(&yabav1.DeletePaymentMethodRequest{Id: "not-a-uuid"}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestPaymentMethodService_NoUserInContextIsUnauthenticatedEvenForInvalidRequests(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, helper.GetTestPool(), uuid.Nil)

	tests := []struct {
		name string
		call func(t *testing.T) connect.Code
	}{
		{
			name: "list",
			call: func(t *testing.T) connect.Code {
				t.Helper()

				_, err := client.ListPaymentMethods(t.Context(),
					connect.NewRequest(&yabav1.ListPaymentMethodsRequest{}))

				return connect.CodeOf(err)
			},
		},
		{
			name: "create with valid request",
			call: func(t *testing.T) connect.Code {
				t.Helper()

				_, err := client.CreatePaymentMethod(t.Context(),
					connect.NewRequest(&yabav1.CreatePaymentMethodRequest{CardTypeId: uuid.NewString()}))

				return connect.CodeOf(err)
			},
		},
		{
			name: "create with non-uuid card_type_id",
			call: func(t *testing.T) connect.Code {
				t.Helper()

				_, err := client.CreatePaymentMethod(t.Context(),
					connect.NewRequest(&yabav1.CreatePaymentMethodRequest{CardTypeId: "not-a-uuid"}))

				return connect.CodeOf(err)
			},
		},
		{
			name: "update with empty request",
			call: func(t *testing.T) connect.Code {
				t.Helper()

				_, err := client.UpdatePaymentMethod(t.Context(),
					connect.NewRequest(&yabav1.UpdatePaymentMethodRequest{}))

				return connect.CodeOf(err)
			},
		},
		{
			name: "delete with non-uuid id",
			call: func(t *testing.T) connect.Code {
				t.Helper()

				_, err := client.DeletePaymentMethod(t.Context(),
					connect.NewRequest(&yabav1.DeletePaymentMethodRequest{Id: "not-a-uuid"}))

				return connect.CodeOf(err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// connect.CodeOf(nil) is CodeUnknown, so this also fails if the call unexpectedly succeeds.
			require.Equal(t, connect.CodeUnauthenticated, test.call(t))
		})
	}
}
