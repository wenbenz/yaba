package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"yaba/internal/database"
	"yaba/internal/email"
	"yaba/internal/handlers"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
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
	if err = waitForConnection(pool); err != nil {
		log.Fatalln("bad db connection:", err)
	}

	log.Println("Connected to db! Applying migrations...")

	if err = applyMigrations(pool); err != nil {
		log.Fatalln(err)
	}

	log.Println("Migrations applied successfully!")

	mailer := buildMailer()

	rootHandler, err := handlers.BuildServerHandler(pool, mailer)
	if err != nil {
		log.Fatalln("could not build root handler:", err)
	}

	// Server setup
	host := os.Getenv("YABA_SERVER_HOST")
	port := os.Getenv("YABA_SERVER_PORT")

	var address string
	if host != "" || port != "" {
		address = host + ":" + port
	}

	yabaServer := http.Server{
		Handler:      rootHandler,
		Addr:         address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on address %q", address) //nolint:gosec // address is from env config, not user input

	err = yabaServer.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}

// buildMailer constructs an SMTPMailer from environment variables, or returns a
// NoopMailer when SMTP_HOST is not set (development / no-email mode).
func buildMailer() email.Mailer {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		log.Println("SMTP_HOST not set — email sending disabled")

		return email.NoopMailer{}
	}

	passwordFile := os.Getenv("SMTP_PASSWORD_FILE")

	var password string

	if passwordFile != "" {
		raw, err := os.ReadFile(passwordFile)
		if err != nil {
			log.Fatalln("could not read SMTP_PASSWORD_FILE:", err)
		}

		password = strings.TrimSpace(string(raw))
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	return email.NewSMTPMailer(email.SMTPConfig{
		Host:     host,
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: password,
		From:     os.Getenv("SMTP_FROM"),
	})
}

func waitForConnection(pool *pgxpool.Pool) error {
	startTime := time.Now()
	ticker := time.NewTicker(time.Second)

	var err error

	for t := range ticker.C {
		err = pool.Ping(context.Background())
		if err == nil {
			ticker.Stop()

			break
		}

		if t.After(startTime.Add(time.Minute)) {
			ticker.Stop()

			return fmt.Errorf("failed to ping database: %w", err)
		}

		log.Println("failed to ping database:", err)
	}

	return nil
}

func applyMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrator: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}
