package models

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Services struct {
	OAuth   OAuthService
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB
}

type ServicesConfig func(*Services) error

func WithGormDB(dialect string, connectionInfo string) ServicesConfig {
	return func(services *Services) error {
		db, err := gorm.Open(postgres.Open(connectionInfo))
		if err != nil {
			return err
		}
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			panic(err)
		}
		services.db = db
		return nil
	}
}

func WithDBLogMode(mode bool) ServicesConfig {
	if mode {
		return func(services *Services) error {
			services.db.Logger = logger.Default.LogMode(logger.Info)
			return nil
		}
	}
	return func(services *Services) error {
		services.db.Logger = logger.Default.LogMode(logger.Silent)
		return nil
	}
}

func WithOAuthService() ServicesConfig {
	return func(services *Services) error {
		services.OAuth = NewOAuthService(services.db)
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
// func (services *Services) Close() error {
// 	return services.db.Close()
// }

func (services *Services) AutoMigrate() error {
	return services.db.AutoMigrate(&User{}, &Gallery{}, &pwReset{}, &OAuth{})
}

func (services *Services) DestructiveReset() error {
	if err := services.db.Migrator().DropTable(&User{}, &Gallery{}, &pwReset{}, &OAuth{}); err != nil {
		return err
	}
	return services.AutoMigrate()
}
