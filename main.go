package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var homeTemplate *template.Template

func home(w http.ResponseWriter, r *http.Request) {
	// set content type
	w.Header().Set("Content-Type", "text/html")
	if err := homeTemplate.Execute(w, nil); err != nil {
		panic(err)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `To get in touch, please send an email to <a 
		href="mailto:support@lenslocked.com">support@lenslocked.com</a>.`)
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
	// load templates
	var err error
	homeTemplate, err = homeTemplate.ParseFiles("views/home.gohtml")
	if err != nil {
		panic(err)
	}

	// create mux router
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.NotFoundHandler = MyNotFound{}

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
