package controllers

import (
	"fmt"
	"go-web-dev/context"
	"go-web-dev/models"
	"go-web-dev/views"
	"log"
	"net/http"
)

func NewGalleryController(galleryService models.GalleryService) *GalleryController {
	return &GalleryController{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		galleryService: galleryService,
	}
}

type GalleryController struct {
	NewView        *views.View
	galleryService models.GalleryService
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
	fmt.Fprintln(w, gallery)
}
