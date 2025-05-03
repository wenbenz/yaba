package database

import (
	"fmt"
	"os"
	"yaba/errors"

	"github.com/Masterminds/squirrel"
)

func init() {
	squirrel.StatementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func GetPGConnectionString() (string, error) {
	missing := []string{}

	pgDatabase := getEnvTrackMissing("POSTGRES_DB", &missing)
	pgUser := getEnvTrackMissing("POSTGRES_USER", &missing)
	pgPasswordFile := getEnvTrackMissing("POSTGRES_PASSWORD_FILE", &missing)
	sslEnabled := getEnvTrackMissing("POSTGRES_SSL_MODE", &missing)
	pgURL := getEnvTrackMissing("POSTGRES_URL", &missing)

	if len(missing) > 0 {
		return "", errors.InvalidStateError{
			Message: fmt.Sprintf("missing postgres env variables: %v", missing),
		}
	}

	pgPassword, err := os.ReadFile(pgPasswordFile)
	if err != nil {
		return "", fmt.Errorf("failed to read postgres password: %w", err)
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		pgUser, pgPassword, pgURL, pgDatabase, sslEnabled), nil
}

func getEnvTrackMissing(key string, missing *[]string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		*missing = append(*missing, key)
	}

	return value
}
