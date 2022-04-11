package main

import (
	"go-web-dev/controllers"
	"go-web-dev/views"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	homeView    *views.View
	contactView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request) {
	must(contactView.Render(w, nil))
}

func main() {
	// load views
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	// init controllers
	userController := controllers.NewUserController()

	// create mux router - routes requests to controllers
	r := mux.NewRouter()
	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/contact", contact).Methods("GET")
	r.HandleFunc("/signup", userController.New).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
