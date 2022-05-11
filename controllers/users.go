package controllers

import (
	"fmt"
	"go-web-dev/models"
	"go-web-dev/views"
	"net/http"
)

// NewUsers is used to create a new Users controller.
// Panics if templates not parsed correctly. Use at setup only.
func NewUserController(userService *models.UserService) *UserController {
	return &UserController{
		NewUserView: views.NewView("bootstrap", "users/new"),
		LoginView:   views.NewView("bootstrap", "users/login"),
		userSerivce: userService,
	}
}

type UserController struct {
	NewUserView *views.View
	LoginView   *views.View
	userSerivce *models.UserService
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
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := userController.userSerivce.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to verify the email and password.
//
// POST /login
func (userController *UserController) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user, err := userController.userSerivce.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			fmt.Fprintln(w, "Invalid email address")
		case models.ErrInvalidPassword:
			fmt.Fprintln(w, "Incorrect password")
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	cookie := http.Cookie{
		Name:  "email",
		Value: user.Email,
	}
	http.SetCookie(w, &cookie)
	fmt.Fprintln(w, user)
}
