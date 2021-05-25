package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

type AuthorsHandler struct {
	as          *models.AuthorService
	ListView    *views.View
	DetailsView *views.View
}

func NewAuthorsHandler(as *models.AuthorService) *AuthorsHandler {
	listView := views.NewView("bootstrap", "templates/authors.gohtml")
	detailsView := views.NewView("bootstrap", "templates/author_details.gohtml")

	return &AuthorsHandler{as, listView, detailsView}
}

type AuthorsViewModel struct {
	Authors []*models.Author
}

type AuthorDetailsViewModel struct {
	Author *models.Author
	Books  []*models.Book
}

func (a *AuthorsHandler) Details(w http.ResponseWriter, r *http.Request) {
	authorId, err := helpers.ParseUintParam(r, "authorId")
	if err != nil {
		http.Error(w, "could not parse request param", http.StatusBadRequest)
		return
	}

	author, err := a.as.GetByID(authorId)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get author by id: %v", err), http.StatusNotFound)
		return
	}

	books, err := a.as.GetBooks(authorId)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get authors books: %v", err), http.StatusInternalServerError)
		return
	}

	a.DetailsView.Render(w, r, &AuthorDetailsViewModel{author, books})
}

func (a *AuthorsHandler) List(w http.ResponseWriter, r *http.Request) {
	authors, err := a.as.GetList()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load Authors: %v", err))
		return
	}
	a.ListView.Render(w, r, AuthorsViewModel{Authors: authors})
}

func (a *AuthorsHandler) Index(w http.ResponseWriter, r *http.Request) {
	authors, err := a.as.GetList()

	if err != nil {
		internalServerError(w, fmt.Sprintf("could not query for authors: %v", err))
		return
	}

	b, _ := json.Marshal(authors)

	w.Write(b)
}
