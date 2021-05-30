package handlers

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

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
	listView := views.NewView("bulma", "templates/authors.gohtml")
	detailsView := views.NewView("bulma", "templates/author_edit.gohtml")

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

func (a *AuthorsHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	authorId, err := helpers.ParseUintParam(r, "authorId")
	if err != nil {
		http.Error(w, "could not parse request param", http.StatusBadRequest)
		return
	}

	var authorEditFormData struct {
		Name string
	}

	err = r.ParseForm()
	if err != nil {
		log.Errorf("could not parse form data: %v", err)
		http.Error(w, "error editing author", http.StatusInternalServerError)
		return
	}

	err = decoder.Decode(&authorEditFormData, r.PostForm)
	if err != nil {
		log.Errorf("could not decode form data: %v", err)
		http.Error(w, "error editing author", http.StatusInternalServerError)
		return
	}

	author, err := a.as.GetByID(authorId)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get author by id: %v", err), http.StatusNotFound)
		return
	}

	author.Name = authorEditFormData.Name

	_, err = a.as.Update(author)
	if err != nil {
		log.Errorf("error updating author: %v", err)
		http.Error(w, "error updating author", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/authors/%d", author.ID), http.StatusFound)
}

func (a *AuthorsHandler) Edit(w http.ResponseWriter, r *http.Request) {
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

func (a *AuthorsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authorId, err := helpers.ParseUintParam(r, "authorId")
	if err != nil {
		http.Error(w, "could not parse request param", http.StatusBadRequest)
		return
	}

	author, err := a.as.GetByID(authorId)
	if err != nil {
		log.Errorf("could not get author: %v", err)
		http.Error(w, "could not find author", http.StatusNotFound)
		return
	}

	if err = a.as.Delete(author); err != nil {
		log.Errorf("could not delete author: %v", err)
		http.Error(w, "error deleting author", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/authors", http.StatusFound)
	return
}

func (a *AuthorsHandler) List(w http.ResponseWriter, r *http.Request) {
	authors, err := a.ba.AuthorsWithBookCount()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load Authors: %v", err))
		return
	}
	a.ListView.Render(w, r, AuthorListViewModel{Authors: authors})
}
