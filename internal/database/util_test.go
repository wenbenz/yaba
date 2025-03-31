package database_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"yaba/internal/database"
)

func TestGetPGConnectionString(t *testing.T) {
	t.Setenv("POSTGRES_DB", "db")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD_FILE", "testdata/password.txt")
	t.Setenv("POSTGRES_URL", "localhost:5432")
	t.Setenv("POSTGRES_SSL_MODE", "disable")

	connectionString, err := database.GetPGConnectionString()
	require.NoError(t, err)
	require.Equal(t, "postgres://user:password1@localhost:5432/db?sslmode=disable", connectionString)
}

func TestGetPGConnectionStringBadPasswordFile(t *testing.T) {
	t.Setenv("POSTGRES_DB", "db")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD_FILE", "testdata/passNoExist")
	t.Setenv("POSTGRES_URL", "localhost:5432")
	t.Setenv("POSTGRES_SSL_MODE", "disable")

	connectionString, err := database.GetPGConnectionString()
	require.Empty(t, connectionString)
	require.ErrorContains(t, err, "failed to read postgres password")
}

func TestGetPGConnectionStringMissigVariable(t *testing.T) {
	t.Parallel()

	connectionString, err := database.GetPGConnectionString()
	require.Empty(t, connectionString)
	require.ErrorContains(t, err, "missing postgres env variables: ")
}
