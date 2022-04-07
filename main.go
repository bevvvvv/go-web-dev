package main

import (
	"fmt"
	"go-web-dev/main/views"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	homeView    *views.View
	contactView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	// set content type
	w.Header().Set("Content-Type", "text/html")
	err := homeView.Template.ExecuteTemplate(w, homeView.Layout, nil)
	if err != nil {
		panic(err)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := contactView.Template.ExecuteTemplate(w, contactView.Layout, nil)
	if err != nil {
		panic(err)
	}
}

// custom NotFoundHandler
type MyNotFound struct{}

func (nf MyNotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<h1>404 Error Not Found</h1>
	<p>This is not the page you were looking for.</p>
	<p>For assistance please <a href="/contact">contact support</a>.</p>`)
}

func main() {
	// load views
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")

	// create mux router
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.NotFoundHandler = MyNotFound{}

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
