package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	r.HandleFunc("/signin", h.signInHandler).Methods("POST")
	r.HandleFunc("/signup", h.signUpHandler).Methods("POST")

	r.HandleFunc("/status", healthcheckHandler).Methods("GET")

	userR := r.PathPrefix("/users/").Subrouter()
	userR.Use(h.authMiddleware)

	userR.HandleFunc("/cards", h.cardsHandler).Methods("GET", "POST")
	userR.HandleFunc("/cards/next", h.cardsNextHandler).Methods("GET", "PUT")

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

type signin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h handler) signInHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var in signin
	json.Unmarshal(body, &in)

	user, err := user.SignIn(h.ctx, h.pool, in.Email, in.Password)
	if err != nil {
		http.Error(w, "email or password incorrect", http.StatusUnauthorized)
		return
	}

	result := make(map[string]string)
	result["jwt"], err = user.GenJWT()
	if err != nil {
		log.Printf("error generating JWT: %s", err)
		http.Error(w, ErrCreatingUser.Error(), http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(result)

	w.Write(body)
}

var ErrCreatingUser = errors.New("error creating the user")

func (h handler) signUpHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	var signup user.SignUp
	json.Unmarshal(body, &signup)

	user, err := signup.Save(h.ctx, h.pool)
	if err != nil {
		log.Printf("error in user signup: %s", err)
		// TODO: Say what the problem is
		http.Error(w, ErrCreatingUser.Error(), http.StatusBadRequest)
		return
	}

	result := make(map[string]string)
	result["jwt"], err = user.GenJWT()
	if err != nil {
		log.Printf("error generating JWT: %s", err)
		http.Error(w, ErrCreatingUser.Error(), http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(result)

	w.Write(body)
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
