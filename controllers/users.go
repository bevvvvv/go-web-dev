package controllers

import (
	"fmt"
	"go-web-dev/models"
	"go-web-dev/rand"
	"go-web-dev/views"
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

func (userController *UserController) New(w http.ResponseWriter, r *http.Request) {
	type Alert struct {
		Level   string
		Message string
	}
	type Data struct {
		Alert Alert
		Yield interface{}
	}
	alert := Alert{
		Level:   "warning",
		Message: "Successfully rendered a dynamic alert!",
	}
	data := Data{
		Alert: alert,
		Yield: "hello!",
	}
	if err := userController.NewUserView.Render(w, data); err != nil {
		panic(err)
	}
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
	if err := userController.userService.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := userController.signIn(w, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user, err := userController.userService.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			fmt.Fprintln(w, "Invalid email address")
		case models.ErrIncorrectPassword:
			fmt.Fprintln(w, "Incorrect password")
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = userController.signIn(w, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
