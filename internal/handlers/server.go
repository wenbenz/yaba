package handlers

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"strings"
)

func BuildServerHandler(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	gqlHandler := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: &Resolver{
		Pool: pool,
	}}))

	mux.Handle("/graphql", gqlHandler)
	mux.Handle("/upload", UploadHandler{
		Pool: pool,
	})

	var handler http.Handler = mux

	singleUserMode := os.Getenv("SINGLE_USER_MODE")
	if strings.ToLower(singleUserMode) == "true" {
		var err error
		if handler, err = InterceptSingleUserMode(handler); err != nil {
			log.Fatalln(err)
		}
	}

	return handler
}
