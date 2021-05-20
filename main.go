package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/wmolicki/bookler/views"

	"github.com/gorilla/csrf"

	"github.com/wmolicki/bookler/handlers"
	"golang.org/x/oauth2"

	"github.com/wmolicki/bookler/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	e := config.NewEnv()
	r := getRouter()

	b := handlers.NewBookHandler(e)
	a := handlers.NewAuthorsHandler(e)

	indexView := views.NewView("bulma", "templates/index.gohtml")

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		indexView.Render(w, r, nil)
	})

	r.Get("/books", b.List)
	r.Get("/books/add", b.Add)
	r.Post("/books/add", b.HandleAdd)
	r.Get("/books/{bookId:[0-9]+}", b.Edit)
	r.Post("/books/{bookId:[0-9]+}", b.HandleEdit)

	r.Get("/authors", a.Display)

	conf := &oauth2.Config{
		ClientID:     "810036611838-k3ur24fbnamqvlu4stsorm47v2onlv0k.apps.googleusercontent.com",
		ClientSecret: "9PD3Fvk7qlraXgKM1T9eQQ0X",
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		},
		RedirectURL: "http://localhost:3333/oauth/google/callback",
	}

	googleRedirect := func(w http.ResponseWriter, r *http.Request) {
		state := csrf.Token(r)
		url := conf.AuthCodeURL(state)

		cookie := http.Cookie{
			Name:     "oidc_state",
			Value:    state,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, url, http.StatusFound)
	}
	r.Get("/oauth/google/connect", googleRedirect)

	googleCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Println(r.RequestURI)
		code := r.FormValue("code")
		state := r.FormValue("state")

		cookie, err := r.Cookie("oidc_state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if cookie.Value != state {
			http.Error(w, "invalid state provided", http.StatusBadRequest)
			return
		}

		ctx := context.TODO()
		token, err := conf.Exchange(ctx, code)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "code: ", code, " state: ", state, " token: %v+", token)
	}
	r.Get("/oauth/google/callback", googleCallback)

	r.Get("/api/v1/books", b.Index)
	r.Post("/api/v/books", b.AddBook)
	r.Patch("/api/v1/books/{bookId:[0-9]+}", b.UpdateBook)

	r.Get("/api/v1/authors", a.Index)
	// r.Get("/authors", a.Index)

	err := http.ListenAndServe(":3333", r)
	if err != nil {
		log.Fatalf("cannot serve: %v", err)
	}
}

func getRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	CSRF := csrf.Protect([]byte("secret-csrf-auth-key-should-be-32-bytes"), csrf.Secure(false))
	r.Use(CSRF)
	return r
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
