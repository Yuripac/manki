package main

import (
	"html/template"
	"io"
	"log"
	"manki/db"
	"manki/handler"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer db.Pool().Close()

	e := echo.New()

	t := &Template{Templates: template.Must(template.ParseGlob("www/*.html"))}
	e.Renderer = t

	// Handle routes
	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/", handler.Home)

	e.GET("/signin", handler.SignInForm)
	e.POST("/signin", handler.SignIn)

	e.GET("/signup", handler.SignUpForm)
	e.POST("/signup", handler.SignUp)

	e.GET("/cards/next", handler.NextCard, handler.AuthMiddleware)
	e.PUT("/cards/next", handler.UpdateNextCard, handler.AuthMiddleware)

	log.Fatal(e.Start(":" + os.Getenv("PORT")))
}
