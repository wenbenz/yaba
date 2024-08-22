package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"strings"
	gqlServer "yaba/graph/server"
	server2 "yaba/graph/server"
	"yaba/internal/handlers"
)

func BuildServerHandler(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	gqlHandler := handler.NewDefaultServer(gqlServer.NewExecutableSchema(gqlServer.Config{Resolvers: &server2.Resolver{
		Pool: pool,
	}}))

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

	return handler
}
