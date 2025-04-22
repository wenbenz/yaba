package helper

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // migration tool
	_ "github.com/golang-migrate/migrate/v4/source/file"       // migration tool
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetTestPool() *pgxpool.Pool {
	return getSingletonPool()
}

func NewIsolatedTestPool() *pgxpool.Pool {
	return initPool(setupTestContainer())
}

//nolint:gochecknoglobals
var getSingletonPool = sync.OnceValue(NewIsolatedTestPool)

// https://golang.testcontainers.org/modules/postgres/#initial-database
func setupTestContainer() *postgres.PostgresContainer {
	ctx := context.Background()

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	//nolint:mnd
	postgresContainer := must(postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	))

	log.Println("started test container")

	connectionString := postgresContainer.MustConnectionString(ctx, "sslmode=disable", "application_name=test")

	migrator := must(migrate.New("file://../../migrations", connectionString))
	if err := migrator.Up(); err != nil {
		log.Fatalf("failed to run migrations migrator: %s", err)
	}

	// cleanup method
	return postgresContainer
}

func initPool(container *postgres.PostgresContainer) *pgxpool.Pool {
	pgxConfig := must(pgxpool.ParseConfig(container.MustConnectionString(context.Background())))

	return must(pgxpool.NewWithConfig(context.Background(), pgxConfig))
}

// This is a convenience func and the generic is to avoid val.(type).
//
//nolint:ireturn
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
