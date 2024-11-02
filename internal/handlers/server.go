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
	mux.Handle("/", http.FileServer(http.Dir(os.Getenv("UI_ROOT_DIR"))))

	var h http.Handler = mux

	h = &auth.SessionInterceptor{
		Pool:        pool,
		Intercepted: h,
	}

	devMode := os.Getenv("DEV_MODE") == "true"
	if devMode {
		h = &corsEnabledHandler{h}
	}

	return h, nil
}

type corsEnabledHandler struct {
	handler http.Handler
}

func (h *corsEnabledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding,"+
			" X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	h.handler.ServeHTTP(w, r)
}
