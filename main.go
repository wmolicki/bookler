package main

import (
	"log"
	"net/http"

	"github.com/wmolicki/bookler/handlers"
	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/middleware"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/oauth"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := getRouter()

	services, err := models.NewServices(
		models.WithDB("sqlite3", "./books_v2.db"),
		models.WithAuthorService(),
		models.WithBookService(),
		models.WithUserService(),
		models.WithOauthConfig(oauth.NewConfig()),
	)
	helpers.Must(err)
	defer services.Close()

	userMiddleware := middleware.NewUserMiddleware(services.User)
	r.Use(userMiddleware.AddUser)

	// static views
	static := handlers.NewStatic()
	r.Method(http.MethodGet, "/", static.Index)
	r.Method(http.MethodGet, "/about", static.About)

	b := handlers.NewBookHandler(services.Author, services.Book)
	a := handlers.NewAuthorsHandler(services.Author)
	u := handlers.NewUserHandler(services.User)

	r.Get("/books", b.List)
	r.Get("/books/add", b.Add)
	r.Post("/books/add", b.HandleAdd)
	r.Get("/books/{bookId:[0-9]+}", b.Edit)
	r.Post("/books/{bookId:[0-9]+}", b.HandleEdit)

	r.Get("/authors", a.Display)

	oh := oauth.NewOauthHandler(services.OauthConfig, services.User)
	r.Get("/oauth/google/connect", oh.SetCookieRedirect)
	r.Get("/oauth/google/callback", oh.Callback)
	r.Post("/tokensignin", oh.TokenSignIn)
	r.Get("/sign_in", u.SignIn)
	r.Post("/sign_out", oh.SignOut)

	r.Get("/api/v1/books", b.Index)
	r.Post("/api/v/books", b.AddBook)
	r.Patch("/api/v1/books/{bookId:[0-9]+}", b.UpdateBook)

	r.Get("/api/v1/authors", a.Index)
	// r.Get("/authors", a.Index)

	err = http.ListenAndServe(":3333", r)
	if err != nil {
		log.Fatalf("cannot serve: %v", err)
	}
}

func getRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMw.Logger)
	r.Use(chiMw.Recoverer)
	r.Use(chiMw.NoCache)

	// TODO: turn this on
	//CSRF := csrf.Protect([]byte("secret-csrf-auth-key-should-be-32-bytes"), csrf.Secure(false))
	//r.Use(CSRF)
	return r
}
