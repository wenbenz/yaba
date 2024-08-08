package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"yaba/handlers"
	"yaba/internal/database"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"
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
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalln("bad db connection:", err)
	}

	log.Println("Connected to db!")

	// Server setup
	server := buildServer(pool)

	err = server.ListenAndServe()
	log.Fatalln("Failed to start server", err)
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
