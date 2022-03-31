package main

import (
	"fmt"
	"net/http"
)

// * denotes a pointer
func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>Welcom to my awesome site!</h1>")
	} else if r.URL.Path == "/contact" {
		fmt.Fprint(w, `To get in touch, please send an email to <a 
		href="mailto:support@lenslocked.com">support@lenslocked.com</a>.`)
	}
}

func main() {
	// routes requests to function
	http.HandleFunc("/", handlerFunc)
	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", nil) // nil uses what is declared above
}
