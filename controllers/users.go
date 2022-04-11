package controllers

import (
	"go-web-dev/views"
	"net/http"
)

// NewUsers is used to create a new Users controller.
// Panics if templates not parsed correctly. Use at setup only.
func NewUserController() *UserController {
	return &UserController{
		NewUserView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

type UserController struct {
	NewUserView *views.View
}

func (u *UserController) New(w http.ResponseWriter, r *http.Request) {
	u.NewUserView.Render(w, nil)
}
