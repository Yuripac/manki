package main

import (
	"context"
	"database/sql"
	"log"
	"manki/config"
	"manki/pkg/handler"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DB = "mysql"
)

func main() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	sc, err := config.NewSecretClient()
	if err != nil {
		log.Fatalf("error inilializing secrets: %s", err)
	}

	s, _ := sc.GetSecret("prod/db_dsn")
	pool, err := sql.Open(DB, s["DB_DSN"])
	if err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer pool.Close()

	server := http.Server{
		Addr:    ":" + "3000",
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
