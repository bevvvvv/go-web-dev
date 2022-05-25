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

	maxMultipartMemory = 100 << 20 // 100 megabytes
)

func NewGalleryController(galleryService models.GalleryService, imageService models.ImageService, r *mux.Router) *GalleryController {
	return &GalleryController{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		IndexView:      views.NewView("bootstrap", "galleries/index"),
		ShowView:       views.NewView("bootstrap", "galleries/show"),
		EditView:       views.NewView("bootstrap", "galleries/edit"),
		galleryService: galleryService,
		imgService:     imageService,
		router:         r,
	}
}

type GalleryController struct {
	NewView        *views.View
	IndexView      *views.View
	ShowView       *views.View
	EditView       *views.View
	galleryService models.GalleryService
	imgService     models.ImageService
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
		galleryController.NewView.Render(w, r, viewData)
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
		galleryController.NewView.Render(w, r, viewData)
		return
	}

	url, err := galleryController.router.Get(ShowGalleryRoute).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries
func (galleryController *GalleryController) Index(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	user := context.User(r.Context())
	galleries, err := galleryController.galleryService.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	viewData.Yield = galleries
	galleryController.IndexView.Render(w, r, viewData)
}

// GET /galleries/:id
func (galleryController *GalleryController) Show(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	gallery, err := galleryController.fetchGallery(w, r)
	viewData.Yield = gallery
	if err != nil {
		return
	}

	viewData.Yield = gallery
	galleryController.ShowView.Render(w, r, viewData)
}

// GET /galleries/:id/edit
func (galleryController *GalleryController) Edit(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	gallery, err := galleryController.fetchGallery(w, r)
	viewData.Yield = gallery
	if err != nil {
		return
	}

	viewData.User = context.User(r.Context())
	if gallery.UserID != viewData.User.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	viewData.Yield = gallery
	galleryController.EditView.Render(w, r, viewData)
}

// POST /galleries/:id/update
func (galleryController *GalleryController) Update(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data
	var form GalleryForm

	gallery, err := galleryController.fetchGallery(w, r)
	viewData.Yield = gallery
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
		galleryController.EditView.Render(w, r, viewData)
		return
	}

	gallery.Title = form.Title
	if err := galleryController.galleryService.Update(gallery); err != nil {
		viewData.SetAlert(err)
		galleryController.EditView.Render(w, r, viewData)
		return
	}

	viewData.Alert = &views.Alert{
		Level:   views.AlertLevelSuccess,
		Message: "Gallery succesfully updated!",
	}
	galleryController.EditView.Render(w, r, viewData)
}

// POST /galleries/:id/images
func (galleryController *GalleryController) Upload(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data

	// get gallery corresponding to path
	gallery, err := galleryController.fetchGallery(w, r)
	viewData.Yield = gallery
	if err != nil {
		return
	}

	// ensure user has access to gallery (owns it)
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	// parse multipart form with images
	err = r.ParseMultipartForm(maxMultipartMemory)
	if err != nil {
		log.Println(err)
		viewData.SetAlert(err)
		galleryController.EditView.Render(w, r, viewData)
	}

	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		srcFile, err := fileHeader.Open()
		if err != nil {
			viewData.SetAlert(err)
			galleryController.EditView.Render(w, r, viewData)
			return
		}
		defer srcFile.Close()

		err = galleryController.imgService.Create(gallery.ID, srcFile, fileHeader.Filename)
		if err != nil {
			viewData.SetAlert(err)
			galleryController.EditView.Render(w, r, viewData)
			return
		}
	}

	viewData.Alert = &views.Alert{
		Level:   views.AlertLevelSuccess,
		Message: "Images succesfully uploaded!",
	}
	galleryController.EditView.Render(w, r, viewData)
}

// POST /galleries/:id/delete
func (galleryController *GalleryController) Delete(w http.ResponseWriter, r *http.Request) {
	var viewData views.Data
	gallery, err := galleryController.fetchGallery(w, r)
	viewData.Yield = gallery
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	if err := galleryController.galleryService.Delete(gallery.ID); err != nil {
		viewData.SetAlert(err)
		viewData.Yield = gallery
		galleryController.EditView.Render(w, r, viewData)
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (galleryController *GalleryController) fetchGallery(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := galleryController.galleryService.ByID(uint(id))
	gallery.Images, _ = galleryController.imgService.ByGalleryID(gallery.ID)
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
