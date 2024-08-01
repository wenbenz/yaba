package helper

import (
	"context"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // migration tool
	_ "github.com/golang-migrate/migrate/v4/source/file"       // migration tool
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// https://golang.testcontainers.org/modules/postgres/#initial-database
func SetupTestContainer() (*postgres.PostgresContainer, func()) {
	ctx := context.Background()

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	//nolint:mnd
	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	log.Println("started test container")

	connectionString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	if err != nil {
		log.Fatalf("failed to retrieve connection string: %s", err)
	}

	migrator, err := migrate.New("file://../../migrations", connectionString)
	if err != nil {
		log.Fatalf("failed to initialize migrator: %s", err)
	}

	if err = migrator.Up(); err != nil {
		log.Fatalf("failed to run migrations migrator: %s", err)
	}

	// cleanup method
	return postgresContainer, func() {
		log.Println("terminating test container")

		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}
}

func SetupTestContainerAndInitPool() (*pgxpool.Pool, func()) {
	container, cleanupFunc := SetupTestContainer()

	pgxConfig, err := pgxpool.ParseConfig(container.MustConnectionString(context.Background()))
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.NewWithConfig(context.TODO(), pgxConfig)
	if err != nil {
		panic(err)
	}

	return pool, cleanupFunc
}
