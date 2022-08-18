package main

import (
	"context"
	"flag"
	"fmt"
	"go-web-dev/controllers"
	"go-web-dev/email"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"go-web-dev/rand"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracer = otel.Tracer("go-web-dev")

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

	configs := make(map[string]*oauth2.Config)
	configs[models.OAuthDropbox] = &oauth2.Config{
		ClientID:     appConfig.Dropbox.ID,
		ClientSecret: appConfig.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  appConfig.Dropbox.AuthURL,
			TokenURL: appConfig.Dropbox.TokenURL,
		},
		RedirectURL: "http://localhost:3000/oauth/dropbox/callback",
	}

	// init controllers
	r := mux.NewRouter()
	staticController := controllers.NewStaticController()
	oauthController := controllers.NewOAuthController(services.OAuth, configs)
	userController := controllers.NewUserController(services.User, emailClient)
	galleriesController := controllers.NewGalleryController(services.Gallery, services.Image, r)

	// login middleware
	userExists := middleware.UserExists{
		UserService: services.User,
	}
	userVerification := middleware.UserVerification{
		UserExists: userExists,
	}

	// opentelemetry middleware
	tp, err := initTracer(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	r.Use(otelmux.Middleware("test-server"))

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
	// dropbox
	r.HandleFunc("/oauth/{service_name:[A-Za-z0-9]+}/connect", userVerification.ApplyFn(oauthController.Connect)).Methods("GET")
	r.HandleFunc("/oauth/{service_name:[A-Za-z0-9]+}/callback", userVerification.ApplyFn(oauthController.Callback)).Methods("GET")
	r.HandleFunc("/oauth/dropbox/test", userVerification.ApplyFn(oauthController.DropboxTest)).Methods("GET")
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

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("35.225.102.64:55681"),
		otlptracehttp.WithInsecure(),
	)
	httpexporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(httpexporter),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}
