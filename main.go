package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"manki/pkg/card"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	pool, err := sql.Open("sqlite3", "/db/manki.db")
	if err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/cards", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "POST":
			body, _ := ioutil.ReadAll(req.Body)

			var c card.Card
			json.Unmarshal(body, &c)

			err := card.Save(ctx, pool, &c)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))

				log.Fatalf("error on card saving: %s", err)
			}

			body, _ = json.Marshal(c)
			w.Write(body)
		case "GET":
			cards, err := card.All(ctx, pool)
			if err != nil {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(err.Error()))

				log.Fatalf("error on cards index: %s", err)
				return
			}

			body, _ := json.Marshal(cards)
			w.Write(body)
		}
	})

	server := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	log.Println("Running server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
