package user_test

import (
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
	"yaba/internal/test/helper"
	"yaba/internal/user"
)

func TestCreateNewUserPasswordHash(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	username := gofakeit.Username()
	password := gofakeit.Password(true, true, true, true, true, 16)

	id, err := user.CreateNewUser(context.Background(), pool, username, password)
	require.NoError(t, err)
	require.NotNil(t, id)
	require.NotEqual(t, *id, uuid.Nil)

	// Make sure the stored password is hashed
	var fetchedPassword []byte

	require.NoError(t, pool.QueryRow(
		context.Background(),
		"SELECT password_hash FROM user_profile WHERE id = $1",
		id).
		Scan(&fetchedPassword))

	passwordHashString := string(fetchedPassword)
	require.NotEqual(t, password, passwordHashString)

	// Make sure the hash is verifiable
	verifiedID, err := user.VerifyUser(context.Background(), pool, username, password)
	require.NoError(t, err)
	require.Equal(t, id, verifiedID)

	// Should fail with wrong password
	shouldBeNil, err := user.VerifyUser(context.Background(), pool, username, passwordHashString)
	require.NoError(t, err)
	require.Nil(t, shouldBeNil)
}