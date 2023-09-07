package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"manki/data"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type handler struct {
	pool *sql.DB
}

func NewRouter(ctx context.Context, pool *sql.DB) http.Handler {
	h := handler{pool}

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
	userAuth, _ := r.Context().Value("user").(*data.User)

	c, err := data.NextCard(r.Context(), h.pool, userAuth.ID)
	if err != nil {
		log.Printf("error searching for next card: %s", err)
		http.Error(w, "no card to remember", http.StatusNotFound)
		return
	}

	if r.Method == "PUT" {
		var params struct {
			Score int8 `json:"score"`
		}
		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &params)
		if err != nil {
			log.Panicf("error on json unmarshal: %s", err)
		}

		if err = data.UpdateMemo(r.Context(), h.pool, c, params.Score); err != nil {
			log.Printf("error updating the next card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	body, _ := json.Marshal(c)
	w.Write(body)
}

func (h handler) cardsHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)

		var c data.Card
		json.Unmarshal(body, &c)

		c.UserId = userAuth.ID

		if err := data.AddCard(r.Context(), h.pool, &c); err != nil {
			log.Printf("error adding a new card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ = json.Marshal(c)
		w.Write(body)
	case "GET":
		cards, err := data.Cards(r.Context(), h.pool, userAuth.ID)
		if err != nil {
			log.Printf("error listing cards: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ := json.Marshal(cards)
		w.Write(body)
	}
}

func (h handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authValues := strings.Split(authHeader, "Bearer "); len(authValues) == 2 {
			userAuth, err := data.FindUserByJWT(r.Context(), h.pool, authValues[1])
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

func (h handler) signInHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var in data.SignIn
	json.Unmarshal(body, &in)

	user, err := in.Validate(r.Context(), h.pool)
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

	var signup data.SignUp
	json.Unmarshal(body, &signup)

	user, err := signup.Save(r.Context(), h.pool)
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

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
}
