package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"time"
	"yaba/internal/database"
	"yaba/internal/handlers"
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

	rootHandler, err := handlers.BuildServerHandler(pool)
	if err != nil {
		log.Fatalln("could not build root handler:", err)
	}

	// Server setup
	port, ok := os.LookupEnv("YABA_PORT")
	if !ok {
		port = "80"
	}

	yabaServer := http.Server{
		Handler:      rootHandler,
		Addr:         ":" + port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	err = yabaServer.ListenAndServe()
	log.Fatalln("Failed to start server", err)
}
