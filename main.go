package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	fakecontext "go-web-dev/context"
	"go-web-dev/controllers"
	"go-web-dev/email"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"go-web-dev/rand"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

func main() {
	prodFlag := flag.Bool("prod", false, "Set to true in production. This ensures that a config file is provided.")
	envFlag := flag.Bool("dbenv", false, "If true, reads database connection values from environment variables.")
	flag.Parse()

	appConfig := LoadConfig(*prodFlag, *envFlag)

	services, err := models.NewServices(
		models.WithGormDB(appConfig.Database.Dialect(), appConfig.Database.ConnectionString()),
		models.WithDBLogMode(!appConfig.IsProd()),
		models.WithOAuthService(),
		models.WithGalleryService(),
		models.WithUserService(appConfig.Pepper, appConfig.HMACKey),
		models.WithImageService(),
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	// services.DestructiveReset()
	services.AutoMigrate()

	emailClient := email.NewClient(email.WithMailgun(appConfig.Mailgun.APIKey, appConfig.Mailgun.PublicAPIKey, appConfig.Mailgun.Domain))

	// init controllers
	r := mux.NewRouter()
	staticController := controllers.NewStaticController()
	userController := controllers.NewUserController(services.User, emailClient)
	galleriesController := controllers.NewGalleryController(services.Gallery, services.Image, r)

	// login middleware
	userExists := middleware.UserExists{
		UserService: services.User,
	}
	userVerification := middleware.UserVerification{
		UserExists: userExists,
	}

	dropboxOAuthConf := &oauth2.Config{
		ClientID:     appConfig.Dropbox.ID,
		ClientSecret: appConfig.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  appConfig.Dropbox.AuthURL,
			TokenURL: appConfig.Dropbox.TokenURL,
		},
		RedirectURL: "http://localhost:3000/oauth/dropbox/callback",
	}
	dropboxRedirect := func(w http.ResponseWriter, r *http.Request) {
		state := csrf.Token(r)

		cookie := http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		url := dropboxOAuthConf.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}
	r.HandleFunc("/oauth/dropbox/connect", userVerification.ApplyFn(dropboxRedirect)).Methods("GET")
	dropboxCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.FormValue("state")
		cookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if cookie == nil || cookie.Value != state {
			http.Error(w, "invalid state provided", http.StatusBadRequest)
		}
		cookie.Value = ""
		cookie.Expires = time.Now()
		http.SetCookie(w, cookie)

		code := r.FormValue("code")
		token, err := dropboxOAuthConf.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%+v", token)

		user := fakecontext.User(r.Context())

		existing, err := services.OAuth.Find(user.ID, models.OAuthDropbox)
		if err == nil {
			services.OAuth.Delete(existing.ID)
		} else if err != models.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userOAuth := models.OAuth{
			UserID:      user.ID,
			ServiceName: models.OAuthDropbox,
			Token:       *token,
		}
		err = services.OAuth.Create(&userOAuth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	r.HandleFunc("/oauth/dropbox/callback", userVerification.ApplyFn(dropboxCallback)).Methods("GET")

	dropboxQuery := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		path := r.FormValue("path")

		user := fakecontext.User(r.Context())
		userOAuth, err := services.OAuth.Find(user.ID, models.OAuthDropbox)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		token := userOAuth.Token

		data := struct {
			Path string `json:"path"`
		}{
			Path: path,
		}
		dataBytes, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		request, err := http.NewRequest(http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", bytes.NewReader(dataBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		request.Header.Add("Content-Type", "application/json")

		client := dropboxOAuthConf.Client(context.TODO(), &token)
		response, err := client.Do(request)
		if err != nil {
			http.Error(w, err.Error(), response.StatusCode)
			return
		}
		defer response.Body.Close()

		io.Copy(w, response.Body)
	}
	r.HandleFunc("/oauth/dropbox/test", userVerification.ApplyFn(dropboxQuery)).Methods("GET")

	// create mux router - routes requests to controllers
	r.Handle("/", staticController.HomeView).Methods("GET")
	r.Handle("/contact", staticController.ContactView).Methods("GET")
	// users
	r.HandleFunc("/signup", userController.New).Methods("GET")
	r.HandleFunc("/signup", userController.Create).Methods("POST")
	r.Handle("/login", userController.LoginView).Methods("GET")
	r.HandleFunc("/login", userController.Login).Methods("POST")
	r.HandleFunc("/logout", userVerification.ApplyFn(userController.Logout)).Methods("POST")
	r.Handle("/password/forgot", userController.ForgotPasswordView).Methods("GET")
	r.HandleFunc("/password/forgot", userController.InitiateReset).Methods("POST")
	r.HandleFunc("/password/reset", userController.ResetPassword).Methods("GET")
	r.HandleFunc("/password/reset", userController.PerformReset).Methods("POST")
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

	bytes, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMiddleware := csrf.Protect(bytes, csrf.Secure(appConfig.IsProd()))

	fmt.Printf("Starting the server at :%d...", appConfig.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Port), csrfMiddleware(userExists.Apply(r)))
}
