package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"yaba/handlers"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize connection pool
	pool, err := pgxpool.New(context.Background(), getPGConnectionString())
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}

	// Ping postgres server to make sure things are working
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalln("bad db connection:", err)
	}

	log.Println("Connected to db!")

	// Server setup
	server := buildServer(pool)

	err = server.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}

func getPGConnectionString() string {
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

	pgURL := net.JoinHostPort(pgHost, pgPort)

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		pgUser, pgPassword, pgURL, pgDatabase, sslEnabled)
}

func buildServer(pool *pgxpool.Pool) http.Server {
	mux := http.NewServeMux()

	mux.Handle("/upload", handlers.UploadHandler{
		Pool: pool,
	})

	mux.Handle("/graphql", handler.New(&handler.Config{
		Schema: createGraphqlSchema(),
		Pretty: true,
	}))

	var handler http.Handler = mux

	singleUserMode := os.Getenv("SINGLE_USER_MODE")
	if strings.ToLower(singleUserMode) == "true" {
		var err error
		if handler, err = handlers.InterceptSingleUserMode(handler); err != nil {
			log.Fatalln(err)
		}
	}

	port, ok := os.LookupEnv("YABA_PORT")
	if !ok {
		port = "8080"
	}

	return http.Server{
		Handler:      handler,
		Addr:         ":" + port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
}

func createGraphqlSchema() *graphql.Schema {
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
		log.Fatalln("failed to create graphql schema: ", err)
	}

	return &schema
}
