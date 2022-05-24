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

var _ GalleryDB = &galleryGorm{}

type galleryValidator struct {
	GalleryDB
}

func (gValidator *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery,
		gValidator.requireUserID,
		gValidator.requireTitle)
	if err != nil {
		return err
	}
	return gValidator.GalleryDB.Create(gallery)
}

func (gValidator *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery,
		gValidator.requireUserID,
		gValidator.requireTitle)
	if err != nil {
		return err
	}
	return gValidator.GalleryDB.Update(gallery)
}

type galleryValFunc func(*Gallery) error

func runGalleryValFuncs(gallery *Gallery, funcs ...galleryValFunc) error {
	for _, function := range funcs {
		if err := function(gallery); err != nil {
			return err
		}
	}
	return nil
}

func (gValidator *galleryValidator) requireUserID(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return ErrRequiredUserID
	}
	return nil
}

func (gValidator *galleryValidator) requireTitle(gallery *Gallery) error {
	if gallery.Title == "" {
		return ErrRequiredTitle
	}
	return nil
}

type GalleryDB interface {
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error

	ByID(id uint) (*Gallery, error)
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gGorm *galleryGorm) Create(gallery *Gallery) error {
	return gGorm.db.Create(gallery).Error
}

func (gGorm *galleryGorm) Update(gallery *Gallery) error {
	return gGorm.db.Save(gallery).Error
}

func (gGorm *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gGorm.db.Where("id = ?", id)
	err := first(db, &gallery)
	return &gallery, err
}
