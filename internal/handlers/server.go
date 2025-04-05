package handlers

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"yaba/graph/server"
	"yaba/internal/auth"
)

func BuildServerHandler(pool *pgxpool.Pool) (http.Handler, error) {
	mux := http.NewServeMux()

	gqlHandler := handler.New(server.NewExecutableSchema(server.Config{Resolvers: &Resolver{
		Pool: pool,
	}}))
	gqlHandler.AddTransport(transport.GET{})
	gqlHandler.AddTransport(transport.POST{})
	gqlHandler.AddTransport(transport.MultipartForm{})
	gqlHandler.Use(extension.Introspection{})

	mux.Handle("/graphql", auth.NewAuthRequired(gqlHandler))

	mux.Handle("/api/upload", auth.NewAuthRequired(UploadHandler{Pool: pool}))
	mux.Handle("/api/register", auth.NewUserHandler(pool))
	mux.Handle("/api/login", auth.VerifyUserHandler(pool))
	mux.Handle("/api/logout", auth.NewLogoutHandler(pool))

	routeReactPages(mux)

	mux.Handle("/", http.FileServer(http.Dir(os.Getenv("UI_ROOT_DIR"))))

	var h http.Handler = mux

	h = &auth.SessionInterceptor{
		Pool:        pool,
		Intercepted: h,
	}

	return h, nil
}

var reactPaths = []string{
	"/dashboard",
	"/login",
	"/budget",
	"/expenditure",
	"/register",
}

func routeReactPages(mux *http.ServeMux) {
	for _, path := range reactPaths {
		routeReactPage(mux, path)
	}
}

func routeReactPage(mux *http.ServeMux, path string) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, os.Getenv("UI_ROOT_DIR")+"/index.html")
	})
}
