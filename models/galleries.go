package models

import "github.com/jinzhu/gorm"

// Gallery is our image container that visitors view
type Gallery struct {
	gorm.Model
	UserID uint    `gorm:"not null;index"`
	Title  string  `gorm:"not null"`
	Images []Image `gorm:"-"`
}

func (gallery *Gallery) SplitImages(n int) [][]Image {
	result := make([][]Image, n)
	for i := 0; i < n; i++ {
		result[i] = make([]Image, 0)
	}
	for i, img := range gallery.Images {
		column := i % n
		result[column] = append(result[column], img)
	}
	return result
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

func (gValidator *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	err := runGalleryValFuncs(&gallery,
		gValidator.validateID)
	if err != nil {
		return err
	}
	return gValidator.GalleryDB.Delete(gallery.ID)
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

func (gValidator *galleryValidator) validateID(gallery *Gallery) error {
	if gallery.ID <= 0 {
		return ErrInvalidID
	}
	return nil
}

type GalleryDB interface {
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error

	ByID(id uint) (*Gallery, error)
	ByUserID(userID uint) ([]Gallery, error)
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

func (gGorm *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gGorm.db.Delete(&gallery).Error
}

func (gGorm *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gGorm.db.Where("id = ?", id)
	err := first(db, &gallery)
	return &gallery, err
}

func (gGorm *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	err := gGorm.db.Where("user_id = ?", userID).Find(&galleries).Error
	return galleries, err
}
