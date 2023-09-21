package handler

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"manki/data"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	ErrCreatingUser = errors.New("error creating the user")
)

func NewRouter(ctx context.Context) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/signin", SignInPageHandler).Methods("GET")
	r.HandleFunc("/signin", SignInHandler).Methods("POST")

	r.HandleFunc("/signup", SignUpHandler).Methods("GET")
	r.HandleFunc("/signup", SignUpHandler).Methods("POST")
	r.HandleFunc("/status", HealthcheckHandler).Methods("GET")

	r.HandleFunc("/", HomeHandler).Methods("GET")

	userR := r.PathPrefix("/users/").Subrouter()
	userR.Use(AuthMiddleware)

	userR.HandleFunc("/cards", CardCreateHandler).Methods("POST")

	userR.HandleFunc("/cards/next", CardNextHandler).Methods("GET")
	userR.HandleFunc("/cards/next", CardUpdateNextHandler).Methods("PUT")

	return r
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := authenticate(r)
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "user", user))

		next.ServeHTTP(w, r)
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/users/cards/next", http.StatusMovedPermanently)
}

func CardNextHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

	c, err := data.NextCard(r.Context(), userAuth.ID)
	if err != nil {
		log.Printf("error searching for next card: %s", err)
		http.Error(w, "no card to remember", http.StatusNotFound)
		return
	}

	t, _ := template.ParseFiles("www/cards/next.html")
	t.Execute(w, c)
}

func CardUpdateNextHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

	c, err := data.NextCard(r.Context(), userAuth.ID)
	if err != nil {
		log.Printf("error searching for next card: %s", err)

		http.Error(w, "no card to remember", http.StatusNotFound)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var params struct {
		Score int8 `json:"score"`
	}
	err = json.Unmarshal(body, &params)
	if err != nil {
		log.Printf("error on json unmarshal: %s", err)

		http.Error(w, "something wrong with the score", http.StatusBadRequest)
		return
	}

	if err = data.Prepare(r.Context(), data.SMemo{}, c, params.Score); err != nil {
		log.Printf("error updating the next card: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, _ = json.Marshal(c)
	w.Write(body)
}

func CardCreateHandler(w http.ResponseWriter, r *http.Request) {
	userAuth, _ := r.Context().Value("user").(*data.User)

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
}

func SignInPageHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := authenticate(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}

	t, _ := template.ParseFiles("www/signin.html")
	t.Execute(w, nil)
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	in := data.SignIn{
		Email: r.Form.Get("email"),
		Psw:   r.Form.Get("password"),
	}

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

	http.SetCookie(w, &http.Cookie{Name: "token", Value: result["jwt"]})

	http.Redirect(w, r, "/users/cards/next", http.StatusMovedPermanently)
}

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

func authenticate(r *http.Request) (*data.User, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}

	userAuth, err := data.FindUserByJWT(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return userAuth, nil
}
