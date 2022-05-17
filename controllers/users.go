package controllers

import (
	"fmt"
	"go-web-dev/models"
	"go-web-dev/rand"
	"go-web-dev/views"
	"log"
	"net/http"
)

// NewUsers is used to create a new Users controller.
// Panics if templates not parsed correctly. Use at setup only.
func NewUserController(userService models.UserService) *UserController {
	return &UserController{
		NewUserView: views.NewView("bootstrap", "users/new"),
		LoginView:   views.NewView("bootstrap", "users/login"),
		userService: userService,
	}
}

type UserController struct {
	NewUserView *views.View
	LoginView   *views.View
	userService models.UserService
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
	var viewData views.Data
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		viewData.SetAlert(err)
		userController.NewUserView.Render(w, viewData)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := userController.userService.Create(&user); err != nil {
		viewData.SetAlert(err)
		userController.NewUserView.Render(w, viewData)
		return
	}
	err := userController.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to verify the email and password.
//
// POST /login
func (userController *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		viewData.SetAlert(err)
		userController.LoginView.Render(w, viewData)
		return
	}

	user, err := userController.userService.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			viewData.AlertError("Invalid email address")
		default:
			viewData.SetAlert(err)
		}
		userController.LoginView.Render(w, viewData)
		return
	}

	err = userController.signIn(w, user)
	if err != nil {
		viewData.SetAlert(err)
		userController.LoginView.Render(w, viewData)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

func (userController *UserController) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = userController.userService.Update(user)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (userController *UserController) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := userController.userService.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "Email is: ", user.Email)
}
