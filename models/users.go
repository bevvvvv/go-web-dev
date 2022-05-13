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
			hmac:   hash.NewHMAC(hmacSecretKey),
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
	hmac hash.HMAC
}

func (uValidator *userValidator) ByRemember(remember string) (*User, error) {
	user := User{
		Remember: remember,
	}
	if err := runUserValFuncs(&user,
		uValidator.hashRemember); err != nil {
		return nil, err
	}
	return uValidator.UserDB.ByRemember(user.RememberHash)
}

func (uValidator *userValidator) Create(user *User) error {
	if err := runUserValFuncs(user,
		uValidator.hashPassword,
		uValidator.generateRemember,
		uValidator.hashRemember); err != nil {
		return err
	}

	return uValidator.UserDB.Create(user)
}

func (uValidator *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user,
		uValidator.hashPassword,
		uValidator.hashRemember); err != nil {
		return err
	}

	return uValidator.UserDB.Update(user)
}

func (uValidator *userValidator) Delete(id uint) error {
	user := User{
		Model: gorm.Model{
			ID: id,
		},
	}
	if err := runUserValFuncs(&user,
		uValidator.validateID); err != nil {
		return err
	}
	return uValidator.UserDB.Delete(id)
}

type userValFunc func(*User) error

func runUserValFuncs(user *User, funcs ...userValFunc) error {
	for _, function := range funcs {
		if err := function(user); err != nil {
			return err
		}
	}
	return nil
}

func (uValidator *userValidator) hashPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uValidator *userValidator) hashRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uValidator.hmac.Hash(user.Remember)
	return nil
}

func (uValidator *userValidator) generateRemember(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uValidator *userValidator) validateID(user *User) error {
	if user.ID <= 0 {
		return ErrInvalidID
	}
	return nil
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
	db *gorm.DB
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db,
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

func (uGorm *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
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
	return uGorm.db.Create(user).Error
}

func (uGorm *userGorm) Update(user *User) error {
	return uGorm.db.Save(user).Error
}

func (uGorm *userGorm) Delete(id uint) error {
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
