package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"manki/pkg/card"
	"manki/pkg/user"
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

	mux.HandleFunc("/cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, _ := ioutil.ReadAll(r.Body)

			var c card.Card
			json.Unmarshal(body, &c)

			if ok, _ := user.Exists(ctx, pool, c.UserId); !ok {
				http.Error(w, "user was not found", http.StatusNotFound)
				return
			}

			if err := card.Save(ctx, pool, &c); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			body, _ = json.Marshal(c)
			w.Write(body)
		case "GET":
			cards, err := card.All(ctx, pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
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

	go func() {
		log.Println("Running server...")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error running server: %s", err)
		}
	}()

	<-ctx.Done()
}
