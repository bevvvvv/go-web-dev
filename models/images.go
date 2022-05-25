package models

import (
	"fmt"
	"io"
	"os"
)

const (
	imageDir = "images/galleries/"
)

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error

	// ByGalleryID(galleryID uint) []string
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

func (imgService *imageService) imagePath(galleryID uint) (string, error) {
	galleryPath := fmt.Sprintf("%s/%v/", imageDir, galleryID)
	return galleryPath, os.MkdirAll(galleryPath, 0755)
}
