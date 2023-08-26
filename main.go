package main

import (
	"context"
	"database/sql"
	"log"
	"manki/pkg/handler"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DB = "mysql"
)

func main() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	dsn := os.Getenv("DB_DSN")
	pool, err := sql.Open(DB, dsn)
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
