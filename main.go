package main

import (
	"context"
	"log"
	"manki/db"
	"manki/handler"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := db.Init(); err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer db.Pool().Close()

	server := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: handler.NewRouter(ctx),
	}

	log.Println("Running server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
