package handlers

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"yaba/graph/server"
	"yaba/internal/auth"
	"yaba/ui"
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

	mux.Handle("/api/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	mux.Handle("/api/register", auth.CreateNewUserHandler(pool))
	mux.Handle("/api/login", auth.VerifyUserHandler(pool))
	mux.Handle("/api/logout", auth.NewLogoutHandler(pool))

	mux.Handle("/api/export/expenditure", auth.NewAuthRequired(exportExpenditureHandler(pool)))

	mux.Handle("/api/", http.NotFoundHandler())

	assets, err := ui.Assets()
	if err != nil {
		return nil, fmt.Errorf("could not load ui assets: %w", err)
	}

	mux.Handle("/", ui.Handler(assets))

	var h http.Handler = mux

	h = &auth.SessionInterceptor{
		Pool:        pool,
		Intercepted: h,
	}

	return h, nil
}
