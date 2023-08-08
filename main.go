package main

import (
	"context"
	"database/sql"
	"log"
	"manki/pkg/handler"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DbDriverName = "sqlite3"
)

func main() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	pool, err := sql.Open(os.Getenv("DB_DRIVER_NAME"), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer pool.Close()

	server := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: handler.New(ctx, pool),
	}

	go func() {
		log.Println("Running server...")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error running server: %s", err)
		}
	}()

	<-ctx.Done()
}
