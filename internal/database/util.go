package database

import (
	"fmt"
	"net"
	"os"
	"yaba/internal/errors"
)

func getEnvTrackMissing(key string, missing *[]string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		*missing = append(*missing, key)
	}

	return value
}

func GetPGConnectionString() (string, error) {
	missing := []string{}

	pgDatabase := getEnvTrackMissing("POSTGRES_DB", &missing)
	pgUser := getEnvTrackMissing("POSTGRES_USER", &missing)
	pgPassword := getEnvTrackMissing("POSTGRES_PASSWORD", &missing)
	pgHost := getEnvTrackMissing("POSTGRES_HOST", &missing)
	pgPort := getEnvTrackMissing("POSTGRES_PORT", &missing)
	sslEnabled := getEnvTrackMissing("POSTGRES_SSL_MODE", &missing)

	if len(missing) > 0 {
		return "", errors.InvalidStateError{
			Message: fmt.Sprintf("missing postgres env variables: %v", missing),
		}
	}

	pgURL := net.JoinHostPort(pgHost, pgPort)

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		pgUser, pgPassword, pgURL, pgDatabase, sslEnabled), nil
}
