package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	Name     string
	Age      float64
	Sports   []string
	Children map[string]string
}

func home(w http.ResponseWriter, r *http.Request) {
	// set content type
	w.Header().Set("Content-Type", "text/html")

	// load template
	t, err := template.ParseFiles("templates/hello.gohtml")
	if err != nil {
		panic(err)
	}

	// send data
	data := User{
		Name:     "John Smith",
		Age:      34.5,
		Sports:   []string{"Basketball", "Volleyball", "Baseball"},
		Children: map[string]string{"Daughter": "Lisa", "Son": "Bradley"},
	}
	t.Execute(w, data)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `To get in touch, please send an email to <a 
		href="mailto:support@lenslocked.com">support@lenslocked.com</a>.`)
}

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<h1>Frequently Asked Questions</h1>
	<h3>Did you write this?</h3>
	<p>Yes I did.</p>`)
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

// http.HandlerFunc(function here) - can be used instead of defining custom class.

func main() {
	// create mux router
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq", faq)
	r.NotFoundHandler = MyNotFound{}

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
