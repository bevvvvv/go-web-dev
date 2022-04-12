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
func (thisUserController *UserController) New(w http.ResponseWriter, r *http.Request) {
	thisUserController.NewUserView.Render(w, nil)
}

type SignupForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form.
// Runs when a user submits the form.
//
// POST /signup
func (u *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	fmt.Fprintln(w, form.Email)
	fmt.Fprintln(w, form.Password)
}
