package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided was invalid")
	ErrInvalidPassword = errors.New("models: Incorrect password provided")
)

type UserService struct {
	db *gorm.DB
}

const userPwPepper = "8#yQhWB$adFN"

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

func (userService *UserService) Authenticate(email, password string) (*User, error) {
	user, err := userService.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	case nil:
		return user, nil
	default:
		return nil, err
	}
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
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
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

func (userService *UserService) AutoMigrate() error {
	return userService.db.AutoMigrate(&User{}).Error
}

func (userService *UserService) DestructiveReset() error {
	if err := userService.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return userService.AutoMigrate()
}

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}
