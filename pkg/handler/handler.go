package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"manki/pkg/card"
	"manki/pkg/user"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type handler struct {
	ctx  context.Context
	pool *sql.DB
}

func NewRouter(ctx context.Context, pool *sql.DB) http.Handler {
	h := handler{ctx, pool}

	r := mux.NewRouter()

	r.Use(h.authMiddleware)

	r.HandleFunc("/status", healthcheckHandler).Methods("GET")
	r.HandleFunc("/cards", h.cardsHandler).Methods("GET", "POST")
	r.HandleFunc("/cards/next", h.cardsNextHandler).Methods("GET", "PUT")

	return r
}

func (h handler) cardsNextHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*user.User)

	c, err := card.Next(h.ctx, h.pool, userAuth.Id)
	if err != nil {
		log.Printf("error searching for next card: %s", err)
		http.Error(w, "no card to remember", http.StatusNotFound)
		return
	}

	if r.Method == "PUT" {
		if err = card.UpdateMemo(h.ctx, h.pool, c, 3); err != nil {
			log.Printf("error updating the next card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	body, _ := json.Marshal(c)
	w.Write(body)
}

func (h handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authValues := strings.Split(authHeader, "Bearer "); len(authValues) == 2 {
			userAuth, err := user.FindByJWT(h.ctx, h.pool, authValues[1])
			if err != nil {
				log.Printf("error on find by jwt: %s", err)

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), "user", userAuth))

			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func (h handler) cardsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, pool := h.ctx, h.pool
	userAuth, _ := r.Context().Value("user").(*user.User)

	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)

		var c card.Card
		json.Unmarshal(body, &c)

		c.UserId = userAuth.Id

		if err := card.Add(ctx, pool, &c); err != nil {
			log.Printf("error adding a new card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ = json.Marshal(c)
		w.Write(body)
	case "GET":
		cards, err := card.All(ctx, pool, userAuth.Id)
		if err != nil {
			log.Printf("error listing cards: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ := json.Marshal(cards)
		w.Write(body)
	}
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
}
