package handlers

import (
	"net/http"

	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

func NewUserHandler(us *models.UserService) *UserHandler {
	signInView := views.NewView("bootstrap", "templates/sign_in.gohtml")

	return &UserHandler{us: us, signIn: signInView}
}

type UserHandler struct {
	us     *models.UserService
	signIn *views.View
}

func (uh *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	uh.signIn.Render(w, r, nil)
}
