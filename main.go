package main

import (
	"go-web-dev/controllers"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// init controllers
	staticController := controllers.NewStaticController()
	userController := controllers.NewUserController()

	// create mux router - routes requests to controllers
	r := mux.NewRouter()
	r.Handle("/", staticController.HomeView).Methods("GET")
	r.Handle("/contact", staticController.ContactView).Methods("GET")
	r.HandleFunc("/signup", userController.New).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
