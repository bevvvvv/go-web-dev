package main

import (
	"fmt"
	"go-web-dev/controllers"
	"go-web-dev/models"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	host     = "host.docker.internal"
	port     = 5432
	user     = "postgres"
	password = "secretpass"
	dbname   = "fakeoku"
)

func main() {
	connectionInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	services, err := models.NewServices(connectionInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	// services.DestructiveReset()
	services.AutoMigrate()

	// init controllers
	staticController := controllers.NewStaticController()
	userController := controllers.NewUserController(services.User)

	// create mux router - routes requests to controllers
	r := mux.NewRouter()
	r.Handle("/", staticController.HomeView).Methods("GET")
	r.Handle("/contact", staticController.ContactView).Methods("GET")
	r.Handle("/signup", userController.NewUserView).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")
	r.Handle("/login", userController.LoginView).Methods("GET")
	r.HandleFunc("/login", userController.Login).Methods("POST")
	r.HandleFunc("/cookietest", userController.CookieTest).Methods("GET")

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
