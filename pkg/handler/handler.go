package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"manki/pkg/card"
	"manki/pkg/user"
	"net/http"
)

type handler struct {
	ctx  context.Context
	pool *sql.DB
}

func New(ctx context.Context, pool *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	h := handler{ctx, pool}

	mux.HandleFunc("/cards", h.cardsHandler)
	mux.HandleFunc("/cards/next", h.cardsNextHandler)

	return mux
}

func (h handler) cardsNextHandler(w http.ResponseWriter, r *http.Request) {
	c, err := card.Next(h.ctx, h.pool)
	if err != nil {
		http.Error(w, "no card to remember", http.StatusNotFound)
		return
	}

	if r.Method == "PUT" {
		if err = card.UpdateMemo(h.ctx, h.pool, c, 3); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	body, _ := json.Marshal(c)
	w.Write(body)
}

func (h handler) cardsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, pool := h.ctx, h.pool

	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)

		var c card.Card
		json.Unmarshal(body, &c)

		if ok, _ := user.Exists(ctx, pool, c.UserId); !ok {
			http.Error(w, "user was not found", http.StatusNotFound)
			return
		}

		if err := card.Add(ctx, pool, &c); err != nil {
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
}
