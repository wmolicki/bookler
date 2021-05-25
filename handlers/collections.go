package handlers

import (
	"fmt"
	"net/http"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

func NewCollectionsHandler(us *models.UserService, bs *models.BookService, cs *models.CollectionsService) *CollectionsHandler {
	listView := views.NewView("bootstrap", "templates/collections.gohtml")

	bs.DestructiveReset()
	return &CollectionsHandler{bs, us, cs, listView}
}

type CollectionsHandler struct {
	bs       *models.BookService
	us       *models.UserService
	cs       *models.CollectionsService
	listView *views.View
}

type CollectionsListViewModel struct {
	Collections []*models.Collection
}

func (c *CollectionsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	collections, err := c.cs.List(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not load collections: %v", err), http.StatusInternalServerError)
		return
	}
	c.listView.Render(w, r, &CollectionsListViewModel{collections})
}
