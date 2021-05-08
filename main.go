package main

import (
	"log"
	"net/http"

	"github.com/wmolicki/bookler/handlers"

	"github.com/wmolicki/bookler/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	e := config.NewEnv()
	r := getRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root."))
	})

	a := handlers.NewAuthorsHandler(e)
	b := handlers.NewBookHandler(e)

	r.Get("/books", b.Index)
	r.Post("/books", b.AddBook)
	r.Patch("/books/{bookId:[0-9]+}", b.UpdateBook)

	r.Get("/authors", a.Index)

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
	return r
}
