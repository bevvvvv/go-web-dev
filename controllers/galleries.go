package controllers

import (
	"go-web-dev/context"
	"go-web-dev/models"
	"go-web-dev/views"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	ShowGalleryRoute = "show_gallery"
)

func NewGalleryController(galleryService models.GalleryService, r *mux.Router) *GalleryController {
	return &GalleryController{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		ShowView:       views.NewView("bootstrap", "galleries/show"),
		EditView:       views.NewView("bootstrap", "galleries/edit"),
		galleryService: galleryService,
		router:         r,
	}
}

type GalleryController struct {
	NewView        *views.View
	ShowView       *views.View
	EditView       *views.View
	galleryService models.GalleryService
	router         *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

func (galleryController *GalleryController) Create(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		viewData.SetAlert(err)
		galleryController.NewView.Render(w, viewData)
		return
	}

	// grab user from request context
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := galleryController.galleryService.Create(&gallery); err != nil {
		viewData.SetAlert(err)
		galleryController.NewView.Render(w, viewData)
		return
	}

	url, err := galleryController.router.Get(ShowGalleryRoute).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		// TODO make this go to the index page (for galleries)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries/:id
func (galleryController *GalleryController) Show(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	gallery, err := galleryController.fetchGallery(w, r)
	if err != nil {
		return
	}

	viewData.Yield = gallery
	galleryController.ShowView.Render(w, viewData)
}

// GET /galleries/:id/edit
func (galleryController *GalleryController) Edit(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	gallery, err := galleryController.fetchGallery(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	viewData.Yield = gallery
	galleryController.EditView.Render(w, viewData)
}

// POST /galleries/:id/update
func (galleryController *GalleryController) Update(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data
	var form GalleryForm

	gallery, err := galleryController.fetchGallery(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		viewData.SetAlert(err)
		galleryController.EditView.Render(w, viewData)
		return
	}

	gallery.Title = form.Title
	if err := galleryController.galleryService.Update(gallery); err != nil {
		viewData.SetAlert(err)
		galleryController.EditView.Render(w, viewData)
		return
	}

	url, err := galleryController.router.Get(ShowGalleryRoute).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		// TODO make this go to the index page (for galleries)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (galleryController *GalleryController) fetchGallery(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := galleryController.galleryService.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, err
		default:
			http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
			return nil, err
		}
	}
	return gallery, nil
}
