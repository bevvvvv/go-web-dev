package main

import (
	"fmt"
	"net/http"
)

// * denotes a pointer
func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Welcom to my awesome site!</h1>")
}

func main() {
	// routes requests to function
	http.HandleFunc("/", handlerFunc)
	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", nil) // nil uses what is declared above
}
