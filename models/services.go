package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB
}

type ServicesConfig func(*Services) error

func WithGormDB(dialect string, connectionInfo string) ServicesConfig {
	return func(services *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		services.db = db
		return nil
	}
}

func WithDBLogMode(mode bool) ServicesConfig {
	return func(services *Services) error {
		services.db.LogMode(mode)
		return nil
	}
}

func WithGalleryService() ServicesConfig {
	return func(services *Services) error {
		services.Gallery = NewGalleryService(services.db)
		return nil
	}
}

func WithUserService(pepper string, hmacSecretKey string) ServicesConfig {
	return func(services *Services) error {
		services.User = NewUserService(services.db, pepper, hmacSecretKey)
		return nil
	}
}

func WithImageService() ServicesConfig {
	return func(services *Services) error {
		services.Image = NewImageService()
		return nil
	}
}

func NewServices(configs ...ServicesConfig) (*Services, error) {
	var services Services
	for _, config := range configs {
		if err := config(&services); err != nil {
			return nil, err
		}
	}
	return &services, nil
}

// Closes the uGorm database connection
func (services *Services) Close() error {
	return services.db.Close()
}

func (services *Services) AutoMigrate() error {
	return services.db.AutoMigrate(&User{}, &Gallery{}, &pwReset{}).Error
}

func (services *Services) DestructiveReset() error {
	if err := services.db.DropTableIfExists(&User{}, &Gallery{}, &pwReset{}).Error; err != nil {
		return err
	}
	return services.AutoMigrate()
}
