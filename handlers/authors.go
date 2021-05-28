package handlers

import (
	"fmt"
	"net/http"

	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

type AuthorsHandler struct {
	as *models.AuthorService
	ba *models.BookAuthorService

	ListView    *views.View
	DetailsView *views.View
}

func NewAuthorsHandler(as *models.AuthorService, ba *models.BookAuthorService) *AuthorsHandler {
	listView := views.NewView("bootstrap", "templates/authors.gohtml")
	detailsView := views.NewView("bootstrap", "templates/author_details.gohtml")

	return &AuthorsHandler{as, ba, listView, detailsView}
}

type AuthorListViewModel struct {
	Authors []*models.AuthorWithBookCount
}

type AuthorDetailsViewModel struct {
	Author    *models.Author
	Books     []*models.Book
	BookCount int
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

	books, err := a.ba.AuthorBooks(author.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get authors books: %v", err), http.StatusInternalServerError)
		return
	}

	a.DetailsView.Render(w, r, &AuthorDetailsViewModel{author, books, len(books)})
}

func (a *AuthorsHandler) List(w http.ResponseWriter, r *http.Request) {
	authors, err := a.ba.AuthorsWithBookCount()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load Authors: %v", err))
		return
	}
	a.ListView.Render(w, r, AuthorListViewModel{Authors: authors})
}
