package main

import (
	"go-web-dev/main/views"
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

	// create mux router
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
