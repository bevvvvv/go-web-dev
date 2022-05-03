package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	ErrNotFound = errors.New("models: resource not found")
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	return &UserService{
		db: db,
	}, nil
}

// Lookup user by id provided.
// 1 - user, nil - user found
// 2 - nil, ErrNotFound - no user found
// 3 - nil, otherError - db is having issue
func (userService *UserService) ById(id uint) (*User, error) {
	var user User
	err := userService.db.Where("id = ?", id).First(&user).Error
	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (userService *UserService) Create(user *User) error {
	return userService.db.Create(user).Error
}

// Closes the UserService database connection
func (userService *UserService) Close() error {
	return userService.db.Close()
}

func (userService *UserService) DestructiveReset() {
	userService.db.DropTableIfExists(&User{})
	userService.db.AutoMigrate(&User{})
}

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:not null;unique_index`
}
