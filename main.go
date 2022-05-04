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
	userService, err := models.NewUserService(connectionInfo)
	if err != nil {
		panic(err)
	}
	defer userService.Close()
	// userService.DestructiveReset()
	userService.AutoMigrate()

	// init controllers
	staticController := controllers.NewStaticController()
	userController := controllers.NewUserController(userService)

	// create mux router - routes requests to controllers
	r := mux.NewRouter()
	r.Handle("/", staticController.HomeView).Methods("GET")
	r.Handle("/contact", staticController.ContactView).Methods("GET")
	r.HandleFunc("/signup", userController.New).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")

	// starts server -- my container exposes 9000 by default
	http.ListenAndServe(":9000", r)
}
