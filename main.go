package main

import (
	"fmt"
	"go-web-dev/controllers"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"go-web-dev/rand"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// TODO add to config
const (
	host     = "host.docker.internal"
	port     = 5432
	user     = "postgres"
	password = "secretpass"
	dbname   = "fakeoku"
	isProd   = false
)

func main() {
	r := mux.NewRouter()

	dbConfig := DefaultPostgresConfig()
	services, err := models.NewServices(dbConfig.Dialect(), dbConfig.ConnectionString())
	if err != nil {
		panic(err)
	}
	defer services.Close()
	// services.DestructiveReset()
	services.AutoMigrate()

	// init controllers
	staticController := controllers.NewStaticController()
	userController := controllers.NewUserController(services.User)
	galleriesController := controllers.NewGalleryController(services.Gallery, services.Image, r)

	// login middleware
	userExists := middleware.UserExists{
		UserService: services.User,
	}
	userVerification := middleware.UserVerification{
		UserExists: userExists,
	}

	// create mux router - routes requests to controllers
	r.Handle("/", staticController.HomeView).Methods("GET")
	r.Handle("/contact", staticController.ContactView).Methods("GET")
	// users
	r.Handle("/signup", userController.NewUserView).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")
	r.Handle("/login", userController.LoginView).Methods("GET")
	r.HandleFunc("/login", userController.Login).Methods("POST")
	// galleries
	r.Handle("/galleries/new", userVerification.Apply(galleriesController.NewView)).Methods("GET")
	r.HandleFunc("/galleries", userVerification.ApplyFn(galleriesController.Index)).Methods("GET")
	r.HandleFunc("/galleries", userVerification.ApplyFn(galleriesController.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).Methods("GET").Name(controllers.ShowGalleryRoute)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", userVerification.ApplyFn(galleriesController.Edit)).Methods("GET").Name(controllers.EditGalleryRoute)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", userVerification.ApplyFn(galleriesController.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", userVerification.ApplyFn(galleriesController.UploadImage)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", userVerification.ApplyFn(galleriesController.DeleteImage)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", userVerification.ApplyFn(galleriesController.Delete)).Methods("POST")
	// serve local image files
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))
	// serve static assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetHandler))

	appConfig := DefaultAppConfig()
	bytes, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMiddleware := csrf.Protect(bytes, csrf.Secure(appConfig.IsProd()))

	fmt.Printf("Starting the server at :%d...", appConfig.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Port), csrfMiddleware(userExists.Apply(r)))
}
