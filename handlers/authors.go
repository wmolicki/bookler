package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

type AuthorsHandler struct {
	as *models.AuthorService
	v  *views.View
}

func NewAuthorsHandler(as *models.AuthorService) *AuthorsHandler {
	view := views.NewView("bulma", "templates/authors.gohtml")

	as.DestructiveReset()
	return &AuthorsHandler{as, view}
}

type AuthorsViewModel struct {
	Authors []*models.Author
}

func (a *AuthorsHandler) Display(w http.ResponseWriter, r *http.Request) {
	authors, err := a.as.GetList()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load Authors: %v", err))
		return
	}
	a.v.Render(w, r, AuthorsViewModel{Authors: authors})
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
