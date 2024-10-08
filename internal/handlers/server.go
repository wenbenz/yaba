package handlers

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"strings"
	"yaba/graph/server"
)

func BuildServerHandler(pool *pgxpool.Pool) (http.Handler, error) {
	mux := http.NewServeMux()

	gqlHandler := handler.NewDefaultServer(server.NewExecutableSchema(server.Config{Resolvers: &Resolver{
		Pool: pool,
	}}))

	mux.Handle("/graphql", gqlHandler)
	mux.Handle("/upload", UploadHandler{
		Pool: pool,
	})
	mux.Handle("/", http.FileServer(http.Dir(os.Getenv("UI_ROOT_DIR"))))

	var handler http.Handler = mux

	singleUserMode := os.Getenv("SINGLE_USER_MODE")
	if strings.ToLower(singleUserMode) == "true" {
		var err error
		if handler, err = InterceptSingleUserMode(handler); err != nil {
			return nil, err
		}
	}

	return handler, nil
}
