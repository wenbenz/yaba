package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"time"
	"yaba/internal/database"
	"yaba/internal/server"
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
	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	for t := range ticker.C {
		err = pool.Ping(context.Background())
		if err == nil || t.After(startTime.Add(time.Minute)) {
			break
		}

		log.Println("failed to ping database:", err)
	}
	ticker.Stop()

	if err != nil {
		log.Fatalln("bad db connection:", err)
	}

	log.Println("Connected to db!")

	// Server setup
	yabaServer := http.Server{
		Handler:      server.BuildServerHandler(pool),
		Addr:         ":9222",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	err = yabaServer.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}
