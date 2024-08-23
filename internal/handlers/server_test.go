package handlers_test

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"yaba/graph/client"
	"yaba/internal/handlers"
	"yaba/internal/test/helper"
)

func TestSingleUserModeServer(t *testing.T) {
	user := uuid.New()

	t.Setenv("SINGLE_USER_MODE", "true")
	t.Setenv("SINGLE_USER_UUID", user.String())

	pool := helper.GetTestPool()
	handler, err := handlers.BuildServerHandler(pool)
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	defer server.Close()

	gql := graphql.NewClient(server.URL+"/graphql", server.Client())
	resp, err := client.CreateBudget(context.Background(), gql, "budget", nil, nil)
	require.NoError(t, err)

	b := resp.GetCreateBudget()
	owner := b.GetOwner()
	require.Equal(t, user.String(), owner)
}

func TestSingleUserModeServerNoUser(t *testing.T) {
	t.Setenv("SINGLE_USER_MODE", "true")

	pool := helper.GetTestPool()
	_, err := handlers.BuildServerHandler(pool)
	require.ErrorContains(t, err, "could not parse UUID from SINGLE_USER_UUID in single user mode")
}
