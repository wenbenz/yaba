package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"yaba/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Read config
	config, err := config.ReadConfig("config/test.yaml")
	if err != nil {
		log.Fatalln("failed to read config %w", err)
	}

	// Initialize connection pool
	_, err = pgxpool.New(context.Background(), config.ConnectionString())
	if err != nil {
		log.Fatalln("failed to connect to database: %w", err)
	}

	// Server setup
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("HELLO")); err != nil {
			log.Println("error writing response")
		}
	})

	server := http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%d", config.Service.Port),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	err = server.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}
