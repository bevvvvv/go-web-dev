package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	imageDir = "images/galleries/"
)

// not stored in db (no GORM)
type Image struct {
	GalleryID uint
	Filename  string
}

func (img *Image) Path() string {
	return fmt.Sprintf("%v%v/%v", imageDir, img.GalleryID, img.Filename)
}

func (img *Image) Route() string {
	urlObject := url.URL{
		Path: "/" + img.Path(),
	}
	return urlObject.String()
}

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	Delete(img *Image) error

	ByGalleryID(galleryID uint) ([]Image, error)
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (imgService *imageService) Create(galleryID uint, srcFile io.Reader, filename string) error {
	galleryPath, err := imgService.imagePath(galleryID)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(galleryPath + filename)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func (imgService *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	galleryPath, _ := imgService.imagePath(galleryID)
	filenames, err := filepath.Glob(galleryPath + "*")
	if err != nil {
		return nil, err
	}
	result := make([]Image, len(filenames))
	for i := range filenames {
		result[i] = Image{
			Filename:  strings.Replace(filenames[i], galleryPath, "", 1),
			GalleryID: galleryID,
		}
	}
	return result, nil
}

func (imgService *imageService) Delete(img *Image) error {
	return os.Remove(img.Path())
}

func (imgService *imageService) imagePath(galleryID uint) (string, error) {
	galleryPath := fmt.Sprintf("%v%v/", imageDir, galleryID)
	return galleryPath, os.MkdirAll(galleryPath, 0755)
}
