package handlers

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"strings"
	"yaba/graph/server"
	"yaba/internal/auth"
)

func BuildServerHandler(pool *pgxpool.Pool) (http.Handler, error) {
	mux := http.NewServeMux()

	gqlHandler := handler.NewDefaultServer(server.NewExecutableSchema(server.Config{Resolvers: &Resolver{
		Pool: pool,
	}}))

	mux.Handle("/graphql", auth.NewAuthRequired(gqlHandler))
	mux.Handle("/upload", auth.NewAuthRequired(UploadHandler{Pool: pool}))
	mux.Handle("/register", auth.NewUserHandler(pool))
	mux.Handle("/login", auth.VerifyUserHandler(pool))
	mux.Handle("/", http.FileServer(http.Dir(os.Getenv("UI_ROOT_DIR"))))

	var handler http.Handler = mux

	singleUserMode := os.Getenv("SINGLE_USER_MODE")
	if strings.ToLower(singleUserMode) == "true" {
		var err error
		if handler, err = InterceptSingleUserMode(handler); err != nil {
			return nil, err
		}
	} else {
		handler = &auth.SessionInterceptor{
			Pool:        pool,
			Intercepted: handler,
		}
	}

	return handler, nil
}
