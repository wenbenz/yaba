package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
	"time"
	"yaba/internal/database"
	"yaba/internal/handlers"
)

func main() {
	// Initialize connection pool
	connectionString, err := database.GetPGConnectionString()
	if err != nil {
		log.Fatalln("could not build connection string:", err)
	}

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}

	// Ping postgres server to make sure things are working
	startTime := time.Now()
	ticker := time.NewTicker(time.Second)

	for t := range ticker.C {
		err = pool.Ping(context.Background())
		if err == nil || t.After(startTime.Add(time.Minute)) {
			ticker.Stop()
			break
		}

		log.Println("failed to ping database:", err)
	}

	if err != nil {
		log.Fatalln("bad db connection:", err)
	}

	log.Println("Connected to db! Applying migrations...")

	db := stdlib.OpenDBFromPool(pool)
	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatalln("could not create postgres driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)

	if err != nil {
		log.Fatalln("could not create migrator:", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalln("could not apply migrations:", err)
	}

	fmt.Println("Migrations applied successfully!")

	rootHandler, err := handlers.BuildServerHandler(pool)
	if err != nil {
		log.Fatalln("could not build root handler:", err)
	}

	// Server setup
	port, ok := os.LookupEnv("YABA_PORT")
	if !ok {
		port = "80"
	}

	yabaServer := http.Server{
		Handler:      rootHandler,
		Addr:         ":" + port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	log.Println("Starting server on port", port)
	err = yabaServer.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}
