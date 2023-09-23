package handler

import (
	"context"
	"log"
	"manki/data"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := auth(c)
		if err != nil {
			return err
		}

		c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), "user", user)))
		return next(c)
	}
}

func Home(c echo.Context) error {
	user, _ := auth(c)
	if user == nil {
		return c.Redirect(http.StatusMovedPermanently, "/signin")
	}

	return c.Redirect(http.StatusMovedPermanently, "/cards/next")
}

func NextCard(c echo.Context) error {
	user, _ := c.Request().Context().Value("user").(*data.User)

	card, err := data.NextCard(c.Request().Context(), user.ID)
	if err != nil {
		log.Printf("error searching for next card: %s", err)
		c.String(http.StatusNotFound, "no card to remember")

		return c.Render(http.StatusBadRequest, "card_next.html", &card)
	}

	return c.Render(http.StatusOK, "card_next.html", &card)
}

func UpdateNextCard(c echo.Context) error {
	user, _ := c.Request().Context().Value("user").(*data.User)

	card, err := data.NextCard(c.Request().Context(), user.ID)
	if err != nil {
		log.Printf("error searching for next card: %s", err)

		return c.Render(http.StatusBadRequest, "card_next.html", card)
	}

	c.Request().ParseForm()
	score, err := strconv.ParseInt(c.FormValue("score"), 10, 0)
	if err := data.Prepare(c.Request().Context(), data.SMemo{}, card, int8(score)); err != nil {
		log.Printf("error updating the next card: %s", err)

		return c.Render(http.StatusBadRequest, "card_next.html", card)
	}

	return c.Render(http.StatusOK, "card_next.html", nil)
}

func CardForm(c echo.Context) error {
	return c.Render(http.StatusOK, "card_new.html", nil)
}

func CreateCard(c echo.Context) error {
	user, _ := c.Request().Context().Value("user").(*data.User)

	c.Request().ParseForm()

	card := &data.Card{
		UserId:   user.ID,
		Sentence: c.FormValue("sentence"),
		Meaning:  c.FormValue("meaning"),
	}

	err := data.AddCard(c.Request().Context(), card)
	if err != nil {
		log.Printf("error adding a new card: %s", err)

		return c.Render(http.StatusBadRequest, "card_new.html", card)
	}

	return c.Render(http.StatusOK, "card_new.html", card)
}

func SignInForm(c echo.Context) error {
	user, _ := auth(c)

	// TODO: Check error instead
	if user != nil {
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	return c.Render(http.StatusOK, "signin.html", nil)
}

func SignIn(c echo.Context) error {
	c.Request().ParseForm()

	signin := data.SignIn{
		Email: c.FormValue("email"),
		Psw:   c.FormValue("password"),
	}

	user, err := signin.Validate(c.Request().Context())
	if err != nil {
		// TODO: Should alert about failure some how
		return c.Render(http.StatusUnauthorized, "signin.html", user)
	}

	jwt, err := user.GenJWT()
	if err != nil {
		log.Printf("error generating JWT: %s", err)
		// TODO: Should alert about failure some how
		return c.Render(http.StatusUnauthorized, "signin.html", user)
	}

	c.SetCookie(&http.Cookie{Name: "token", Value: jwt})

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func SignUpForm(c echo.Context) error {
	user, _ := auth(c)

	// TODO: Check error instead
	if user != nil {
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	return c.Render(http.StatusOK, "signup.html", nil)
}

func SignUp(c echo.Context) error {
	c.Request().ParseForm()

	signup := data.SignUp{
		Name:  c.FormValue("name"),
		Email: c.FormValue("email"),
		Psw:   c.FormValue("password"),
	}

	user, err := signup.Save(c.Request().Context())
	if err != nil {
		log.Printf("error in user signup: %s", err)

		return c.Render(http.StatusBadRequest, "signup.html", nil)
	}

	jwt, err := user.GenJWT()
	if err != nil {
		log.Printf("error generating JWT: %s", err)

		return c.Render(http.StatusBadRequest, "signup.html", nil)
	}

	c.SetCookie(&http.Cookie{Name: "token", Value: jwt})

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func auth(c echo.Context) (*data.User, error) {
	cookie, err := c.Cookie("token")
	if err != nil {
		return nil, err
	}

	userAuth, err := data.FindUserByJWT(c.Request().Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return userAuth, nil
}
