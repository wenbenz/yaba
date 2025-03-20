package handlers

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
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
	mux.Handle("/logout", auth.NewLogoutHandler(pool))
	mux.Handle("/", http.FileServer(http.Dir(os.Getenv("UI_ROOT_DIR"))))

	var h http.Handler = mux

	h = &auth.SessionInterceptor{
		Pool:        pool,
		Intercepted: h,
	}

	return h, nil
}
