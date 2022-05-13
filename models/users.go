package models

import (
	"errors"
	"go-web-dev/hash"
	"go-web-dev/rand"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided was invalid")
	ErrInvalidPassword = errors.New("models: Incorrect password provided")
)

const userPwPepper = "8#yQhWB$adFN"
const hmacSecretKey = "secret-hmac-key"

// User accounts in database
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// UserService is a set of methods used to manipulate and
// work with the user model.
type UserService interface {
	// Authenticate will verify the provided email/pw are correct
	// corresponding user to inputs is returned when correct
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(connectionInfo string) (UserService, error) {
	uGorm, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &userService{
		UserDB: &userValidator{
			UserDB: uGorm,
		},
	}, nil
}

type userService struct {
	UserDB
}

func (uService *userService) Authenticate(email, password string) (*User, error) {
	user, err := uService.ByEmail(email)
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

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
}

// UserDB is used to interact with users table.
//
// Single user queries:
// 1 - user, nil - user found
// 2 - nil, ErrNotFound - no user found
// 3 - nil, otherError - db is having issue
type UserDB interface {
	// Methods for querying single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(RememberToken string) (*User, error)

	//Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Used to close a DB connection
	Close() error

	// Migration Helpers
	AutoMigrate() error
	DestructiveReset() error
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

func (uGorm *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := uGorm.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (uGorm *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := uGorm.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (uGorm *userGorm) ByRemember(remember string) (*User, error) {
	var user User
	rememberHash := uGorm.hmac.Hash(remember)
	db := uGorm.db.Where("remember_hash = ?", rememberHash)
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

func (uGorm *userGorm) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}
	user.RememberHash = uGorm.hmac.Hash(user.Remember)
	return uGorm.db.Create(user).Error
}

func (uGorm *userGorm) Update(user *User) error {
	if user.Password != "" {
		pwBytes := []byte(user.Password + userPwPepper)
		hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(hashedBytes)
		user.Password = ""
	}
	if user.Remember != "" {
		user.RememberHash = uGorm.hmac.Hash(user.Remember)
	}
	return uGorm.db.Save(user).Error
}

func (uGorm *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return uGorm.db.Delete(&user).Error
}

// Closes the uGorm database connection
func (uGorm *userGorm) Close() error {
	return uGorm.db.Close()
}

func (uGorm *userGorm) AutoMigrate() error {
	return uGorm.db.AutoMigrate(&User{}).Error
}

func (uGorm *userGorm) DestructiveReset() error {
	if err := uGorm.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return uGorm.AutoMigrate()
}
