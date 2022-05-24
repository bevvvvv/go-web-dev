package models

import "github.com/jinzhu/gorm"

// Gallery is our image container that visitors view
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Title  string `gorm:"not null"`
}

type GalleryService interface {
	GalleryDB
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{db},
		},
	}
}

type galleryService struct {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}

type GalleryDB interface {
	Create(gallery *Gallery) error
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gGorm *galleryGorm) Create(gallery *Gallery) error {
	return gGorm.db.Create(gallery).Error
}
