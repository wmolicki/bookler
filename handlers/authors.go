package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wmolicki/bookler/config"

	"github.com/wmolicki/bookler/models"
)

type AuthorsHandler struct {
	as *models.AuthorService
}

func NewAuthorsHandler(env *config.Env) *AuthorsHandler {
	as := models.NewAuthorService(env)
	as.DestructiveReset()
	return &AuthorsHandler{as}
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
