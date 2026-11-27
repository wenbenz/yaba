package rpc_test

import (
	"testing"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/test/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newRewardCardClient(t *testing.T, pool *pgxpool.Pool, user uuid.UUID) yabav1connect.RewardCardServiceClient {
	t.Helper()

	srv := serve(t, pool, user)

	return yabav1connect.NewRewardCardServiceClient(srv.Client(), srv.URL+"/rpc")
}

func TestCreateAndListRewardCards(t *testing.T) {
	t.Parallel()

	client := newRewardCardClient(t, helper.GetTestPool(), uuid.New())
	issuer := uuid.NewString() // the catalog is shared by parallel tests

	reqs := []*yabav1.CreateRewardCardRequest{
		{
			Name: "Freedom", Issuer: issuer, Region: "US", RewardType: "cash",
			Categories: []*yabav1.RewardCategory{{Category: "Dining", Rate: 3}},
		},
		{Name: "Sapphire", Issuer: issuer, Region: "CA", RewardType: "points"},
	}
	created := make([]*yabav1.RewardCard, 0, len(reqs))

	for _, req := range reqs {
		res, err := client.CreateRewardCard(t.Context(), connect.NewRequest(req))
		require.NoError(t, err)

		created = append(created, res.Msg.GetRewardCard())
	}

	freedom := &yabav1.RewardCard{
		Id: created[0].GetId(), Name: "Freedom", Issuer: issuer, Region: "US", RewardType: "cash",
		Categories: []*yabav1.RewardCategory{{Category: "Dining", Rate: 3}},
	}
	require.NoError(t, uuid.Validate(freedom.GetId()))
	require.True(t, proto.Equal(freedom, created[0]), "created: %v", created[0])
	require.Empty(t, created[1].GetCategories())

	res, err := client.ListRewardCards(t.Context(), connect.NewRequest(&yabav1.ListRewardCardsRequest{Issuer: issuer}))
	require.NoError(t, err)

	listed := res.Msg.GetRewardCards()
	require.Len(t, listed, 2)
	require.True(t, proto.Equal(freedom, listed[0]), "listed: %v", listed[0])
	require.True(t, proto.Equal(created[1], listed[1]), "listed: %v", listed[1])

	for _, tc := range []struct {
		name string
		req  *yabav1.ListRewardCardsRequest
		want []string
	}{
		{"name prefix", &yabav1.ListRewardCardsRequest{Name: "sap"}, []string{"Sapphire"}},
		{"region prefix", &yabav1.ListRewardCardsRequest{Region: "c"}, []string{"Sapphire"}},
		{"limit", &yabav1.ListRewardCardsRequest{Limit: new(int32(1))}, []string{"Freedom"}},
		{"zero limit", &yabav1.ListRewardCardsRequest{Limit: new(int32(0))}, nil},
		{"offset", &yabav1.ListRewardCardsRequest{Offset: new(int32(1))}, []string{"Sapphire"}},
		{"no match", &yabav1.ListRewardCardsRequest{Name: "none"}, nil},
	} {
		tc.req.Issuer = issuer

		res, err = client.ListRewardCards(t.Context(), connect.NewRequest(tc.req))
		require.NoError(t, err, tc.name)

		var names []string
		for _, card := range res.Msg.GetRewardCards() {
			names = append(names, card.GetName())
		}

		require.Equal(t, tc.want, names, tc.name)
	}
}

func TestRewardCardService_InvalidArgument(t *testing.T) {
	t.Parallel()

	client := newRewardCardClient(t, helper.GetTestPool(), uuid.New())

	_, err := client.CreateRewardCard(t.Context(), connect.NewRequest(
		&yabav1.CreateRewardCardRequest{Issuer: uuid.NewString(), Region: "US", RewardType: "points"}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "empty name")

	for name, req := range map[string]*yabav1.ListRewardCardsRequest{
		"negative limit":  {Limit: new(int32(-1))},
		"negative offset": {Offset: new(int32(-1))},
	} {
		_, err = client.ListRewardCards(t.Context(), connect.NewRequest(req))
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), name)
	}
}

func TestRewardCardService_RequiresUser(t *testing.T) {
	t.Parallel()

	client := newRewardCardClient(t, helper.GetTestPool(), uuid.Nil)

	_, err := client.ListRewardCards(t.Context(), connect.NewRequest(&yabav1.ListRewardCardsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))

	// Auth runs before validation.
	_, err = client.CreateRewardCard(t.Context(), connect.NewRequest(&yabav1.CreateRewardCardRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestRewardCardService_DatabaseErrorIsInternal(t *testing.T) {
	t.Parallel()

	pool, err := pgxpool.NewWithConfig(t.Context(), helper.GetTestPool().Config())
	require.NoError(t, err)
	pool.Close()

	client := newRewardCardClient(t, pool, uuid.New())

	_, err = client.ListRewardCards(t.Context(), connect.NewRequest(&yabav1.ListRewardCardsRequest{}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))

	_, err = client.CreateRewardCard(t.Context(), connect.NewRequest(&yabav1.CreateRewardCardRequest{
		Name: "Freedom", Issuer: uuid.NewString(), Region: "US", RewardType: "cash",
	}))
	require.Equal(t, connect.CodeInternal, connect.CodeOf(err))
}
