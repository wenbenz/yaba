package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"yaba/internal/database"
	"yaba/internal/handlers"

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

	graphqlSchema, err := handlers.CreateGraphqlSchema()
	if err != nil {
		log.Fatalln("failed to create graphql schema: ", err)
	}

	gqlHandler := handler.New(&handler.Config{
		Schema: graphqlSchema,
		Pretty: true,
	})

	mux.Handle("/graphql", gqlHandler)
	mux.Handle("/upload", handlers.UploadHandler{
		Pool: pool,
	})

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
