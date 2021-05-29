package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	chiMw "github.com/go-chi/chi/v5/middleware"

	"github.com/wmolicki/bookler/handlers"
	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/middleware"
	"github.com/wmolicki/bookler/models"
)

func main() {
	r := getRouter()

	services, err := models.NewServices(
		models.WithDB("sqlite3", "./books_v2.db"),
		models.WithAuthorService(),
		models.WithBookService(),
		models.WithBookAuthorService(),
		models.WithUserService(),
		models.WithUserBookService(),
		models.WithOauthConfig(handlers.NewConfig()),
		models.WithCollectionsService(),
	)
	helpers.Must(err)
	defer services.Close()

	userMiddleware := middleware.NewUserMiddleware(services.User)
	r.Use(userMiddleware.AddUser)

	// static views
	static := handlers.NewStatic()
	r.Handle("/", static.Index).Methods(http.MethodGet)
	r.Handle("/about", static.About).Methods(http.MethodGet)

	b := handlers.NewBookHandler(services.Author, services.BookAuthor, services.Book, services.UserBook)
	a := handlers.NewAuthorsHandler(services.Author, services.BookAuthor)
	u := handlers.NewUserHandler(services.User)
	c := handlers.NewCollectionsHandler(services.User, services.Book, services.Collections)

	r.HandleFunc("/collections", c.List).Methods(http.MethodGet)
	r.HandleFunc("/collections/add", c.Add).Methods(http.MethodGet)
	r.HandleFunc("/collections/add", c.HandleAdd).Methods(http.MethodPost)
	r.HandleFunc("/collections/{collectionId:[0-9]+}/delete", c.HandleDelete).Methods(http.MethodPost)
	r.HandleFunc("/collections/{collectionId:[0-9]+}/book/add", c.HandleAddBook).Methods(http.MethodPost)
	r.HandleFunc("/collections/{collectionId:[0-9]+}", c.Edit).Methods(http.MethodGet)
	r.HandleFunc("/collections/{collectionId:[0-9]+}", c.HandleEdit).Methods(http.MethodPost)
	r.HandleFunc("/collections/{collectionId:[0-9]+}/book/{bookId:[0-9]+}/delete", c.HandleBookDelete).Methods(http.MethodPost)

	r.HandleFunc("/books", b.List).Methods(http.MethodGet)
	r.HandleFunc("/books/add", b.Add).Methods(http.MethodGet)
	r.HandleFunc("/books/add", b.HandleAdd).Methods(http.MethodPost)
	r.HandleFunc("/books/{bookId:[0-9]+}", b.Edit).Methods(http.MethodGet)
	r.HandleFunc("/books/{bookId:[0-9]+}", b.HandleEdit).Methods(http.MethodPost)
	r.HandleFunc("/books/{bookId:[0-9]+}/delete", b.HandleDelete).Methods(http.MethodPost)

	r.HandleFunc("/authors", a.List).Methods(http.MethodGet)
	r.HandleFunc("/authors/{authorId:[0-9]+}", a.Details).Methods(http.MethodGet)
	r.HandleFunc("/authors/{authorId:[0-9]+}/delete", a.Delete).Methods(http.MethodPost)

	oh := handlers.NewOauthHandler(services.OauthConfig, services.User)
	r.HandleFunc("/oauth/google/connect", oh.SetCookieRedirect).Methods(http.MethodGet)
	r.HandleFunc("/oauth/google/callback", oh.Callback).Methods(http.MethodGet)
	r.HandleFunc("/tokensignin", oh.TokenSignIn).Methods(http.MethodPost)
	r.HandleFunc("/sign_in", u.SignIn).Methods(http.MethodGet)
	r.HandleFunc("/sign_out", oh.SignOut).Methods(http.MethodPost)

	r.HandleFunc("/api/v1/books", b.Index).Methods(http.MethodGet)
	r.HandleFunc("/api/v/books", b.AddBook).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/books/{bookId:[0-9]+}", b.UpdateBook).Methods(http.MethodPatch)

	staticHandler := http.FileServer(http.Dir("./static"))
	staticHandler = http.StripPrefix("/static/", staticHandler)
	r.PathPrefix("/static").Handler(staticHandler)

	err = http.ListenAndServe(":3333", r)
	if err != nil {
		log.Fatalf("cannot serve: %v", err)
	}
}

func getRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(chiMw.Logger)
	r.Use(chiMw.Recoverer)
	r.Use(chiMw.NoCache)

	// TODO: turn this on
	//CSRF := csrf.Protect([]byte("secret-csrf-auth-key-should-be-32-bytes"), csrf.Secure(false))
	//r.Use(CSRF)
	return r
}
