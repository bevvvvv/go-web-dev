package controllers

import (
	"fmt"
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

// New is used to render the form where
// a user can create a new user account.
//
// GET /signup
func (u *UserController) New(w http.ResponseWriter, r *http.Request) {
	u.NewUserView.Render(w, nil)
}

// Create is used to process the signup form.
// Runs when a user submits the form.
//
// POST /signup
func (u *UserController) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is a temporary response.")
}
