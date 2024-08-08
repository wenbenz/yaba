package database_test

import (
	"testing"
	"yaba/internal/database"

	"github.com/stretchr/testify/require"
)

func TestGetPGConnectionString(t *testing.T) {
	t.Setenv("POSTGRES_DB", "db")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_SSL_MODE", "disable")

	connectionString, err := database.GetPGConnectionString()
	require.NoError(t, err)
	require.Equal(t, "postgres://user:password@localhost:5432/db?sslmode=disable", connectionString)
}

func TestGetPGConnectionStringMissigVariable(t *testing.T) {
	t.Parallel()

	connectionString, err := database.GetPGConnectionString()
	require.Equal(t, "", connectionString)
	require.ErrorContains(t, err, "missing postgres env variables: "+
		"[POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD POSTGRES_HOST POSTGRES_PORT POSTGRES_SSL_MODE]")
}
