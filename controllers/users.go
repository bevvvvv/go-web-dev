package controllers

import (
	"go-web-dev/models"
	"go-web-dev/views"
	"net/http"
)

// NewUsers is used to create a new Users controller.
// Panics if templates not parsed correctly. Use at setup only.
func NewUserController(userService *models.UserService) *UserController {
	return &UserController{
		NewUserView: views.NewView("bootstrap", "users/new"),
		userSerivce: userService,
	}
}

type UserController struct {
	NewUserView *views.View
	userSerivce *models.UserService
}

// New is used to render the form where
// a user can create a new user account.
//
// GET /signup
func (thisUserController *UserController) New(w http.ResponseWriter, r *http.Request) {
	thisUserController.NewUserView.Render(w, nil)
}

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form.
// Runs when a user submits the form.
//
// POST /signup
func (userController *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name:  form.Name,
		Email: form.Email,
	}
	if err := userController.userSerivce.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
