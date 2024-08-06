package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Read config
	pgDatabase, ok1 := os.LookupEnv("POSTGRES_DB")
	pgUser, ok2 := os.LookupEnv("POSTGRES_USER")
	pgPassword, ok3 := os.LookupEnv("POSTGRES_PASSWORD")
	pgHost, ok4 := os.LookupEnv("POSTGRES_HOST")
	pgPort, ok5 := os.LookupEnv("POSTGRES_PORT")
	sslEnabled, ok6 := os.LookupEnv("POSTGRES_SSL_MODE")

	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		log.Fatalln("env missing postgres variables")
	}

	port, ok := os.LookupEnv("YABA_PORT")
	if !ok {
		port = "8080"
	}

	pgURL := net.JoinHostPort(pgHost, pgPort)
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		pgUser, pgPassword, pgURL, pgDatabase, sslEnabled)

	// Initialize connection pool
	_, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalln("failed to connect to database: %w", err)
	}

	// GraphQL Schema
	schema, err := createGraphqlSchema()
	if err != nil {
		log.Fatalln("failed to create new schema: %w", err)
	}

	gqlHandler := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	// Server setup
	mux := http.NewServeMux()
	mux.Handle("/graphql", gqlHandler)

	server := http.Server{
		Handler:      mux,
		Addr:         ":" + port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	err = server.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}

func createGraphqlSchema() (graphql.Schema, error) {
	fields := graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Name: "Ping",
			Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
				return "pong", nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: fields,
	}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return graphql.Schema{}, fmt.Errorf("failed to create graphql schema: %w", err)
	}

	return schema, nil
}
