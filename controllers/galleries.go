package controllers

import (
	"go-web-dev/models"
	"go-web-dev/views"
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
