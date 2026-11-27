package rpc_test

import (
	"strings"
	"testing"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/gen/yaba/v1/yabav1connect"
	"yaba/internal/test/helper"
	"yaba/internal/user"

	"connectrpc.com/connect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// newUserClient creates a real user and returns a client whose requests carry it.
func newUserClient(t *testing.T, username, password string) (yabav1connect.UserServiceClient, uuid.UUID) {
	t.Helper()

	id, err := user.CreateNewUser(t.Context(), helper.GetTestPool(), username, password)
	require.NoError(t, err)

	srv := serve(t, helper.GetTestPool(), *id)

	return yabav1connect.NewUserServiceClient(srv.Client(), srv.URL+"/rpc"), *id
}

func TestGetAndUpdateProfile(t *testing.T) {
	t.Parallel()

	username := gofakeit.Username()
	client, id := newUserClient(t, username, "password")

	res, err := client.GetProfile(t.Context(), connect.NewRequest(&yabav1.GetProfileRequest{}))
	require.NoError(t, err)

	want := &yabav1.Profile{Id: id.String(), Username: username}
	require.True(t, proto.Equal(want, res.Msg.GetProfile()), "profile: %v", res.Msg.GetProfile())

	email := gofakeit.Email()
	updated, err := client.UpdateProfile(t.Context(), connect.NewRequest(&yabav1.UpdateProfileRequest{Email: email}))
	require.NoError(t, err)
	require.Equal(t, email, updated.Msg.GetProfile().GetEmail())

	res, err = client.GetProfile(t.Context(), connect.NewRequest(&yabav1.GetProfileRequest{}))
	require.NoError(t, err)
	require.Equal(t, email, res.Msg.GetProfile().GetEmail())

	_, err = client.UpdateProfile(t.Context(), connect.NewRequest(&yabav1.UpdateProfileRequest{Email: "not-an-email"}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "malformed email")

	other, _ := newUserClient(t, gofakeit.Username(), "password")

	_, err = other.UpdateProfile(t.Context(),
		connect.NewRequest(&yabav1.UpdateProfileRequest{Email: strings.ToUpper(email)}))
	require.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err), "email taken, case-insensitively")
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	username := gofakeit.Username()
	client, id := newUserClient(t, username, "old")
	req := connect.NewRequest(&yabav1.ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "new"})

	_, err := client.ChangePassword(t.Context(), req)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "wrong current password")

	req.Msg.CurrentPassword = "old"

	_, err = client.ChangePassword(t.Context(), req)
	require.NoError(t, err)

	verified, err := user.VerifyUser(t.Context(), helper.GetTestPool(), username, "new")
	require.NoError(t, err)
	require.Equal(t, &id, verified)

	verified, err = user.VerifyUser(t.Context(), helper.GetTestPool(), username, "old")
	require.NoError(t, err)
	require.Nil(t, verified, "old password must no longer verify")
}
