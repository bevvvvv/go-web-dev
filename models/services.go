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

func NewServices(dialect string, connectionInfo string) (*Services, error) {
	// TODO setup config
	db, err := gorm.Open(dialect, connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &Services{
		User:    NewUserService(db),
		Gallery: NewGalleryService(db),
		Image:   NewImageService(),
		db:      db,
	}, nil
}

// // Used to close a DB connection
// Close() error

// // Migration Helpers
// AutoMigrate() error
// DestructiveReset() error

// Closes the uGorm database connection
func (services *Services) Close() error {
	return services.db.Close()
}

func (services *Services) AutoMigrate() error {
	return services.db.AutoMigrate(&User{}, &Gallery{}).Error
}

func (services *Services) DestructiveReset() error {
	if err := services.db.DropTableIfExists(&User{}, &Gallery{}).Error; err != nil {
		return err
	}
	return services.AutoMigrate()
}
