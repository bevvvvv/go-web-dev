package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	ErrNotFound  = errors.New("models: resource not found")
	ErrInvalidID = errors.New("models: ID provided was invalid")
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
	db := userService.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (userService *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := userService.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return ErrNotFound
	default:
		return err
	}
}

func (userService *UserService) Create(user *User) error {
	return userService.db.Create(user).Error
}

func (userService *UserService) Update(user *User) error {
	return userService.db.Save(user).Error
}

func (userService *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return userService.db.Delete(&user).Error
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
