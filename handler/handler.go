package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"manki/data"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func NewRouter(ctx context.Context) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/signin", SignInHandler).Methods("POST")
	r.HandleFunc("/signup", SignUpHandler).Methods("POST")
	r.HandleFunc("/status", HealthcheckHandler).Methods("GET")

	userR := r.PathPrefix("/users/").Subrouter()
	userR.Use(AuthMiddleware)

	userR.HandleFunc("/cards", CardsHandler).Methods("GET", "POST")
	userR.HandleFunc("/cards/next", CardsNextHandler).Methods("GET", "PUT")

	return r
}

func CardsNextHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

	c, err := data.NextCard(r.Context(), userAuth.ID)
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

		if err = data.UpdateMemo(r.Context(), c, params.Score); err != nil {
			log.Printf("error updating the next card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	body, _ := json.Marshal(c)
	w.Write(body)
}

func CardsHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)

		var c data.Card
		json.Unmarshal(body, &c)

		c.UserId = userAuth.ID

		if err := data.AddCard(r.Context(), &c); err != nil {
			log.Printf("error adding a new card: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ = json.Marshal(c)
		w.Write(body)
	case "GET":
		cards, err := data.Cards(r.Context(), userAuth.ID)
		if err != nil {
			log.Printf("error listing cards: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ := json.Marshal(cards)
		w.Write(body)
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authValues := strings.Split(authHeader, "Bearer "); len(authValues) == 2 {
			userAuth, err := data.FindUserByJWT(r.Context(), authValues[1])
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

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var in data.SignIn
	json.Unmarshal(body, &in)

	user, err := in.Validate(r.Context())
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

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	var signup data.SignUp
	json.Unmarshal(body, &signup)

	user, err := signup.Save(r.Context())
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

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
}
