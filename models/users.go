package models

import (
	"go-web-dev/hash"
	"go-web-dev/rand"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

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

func NewUserService(db *gorm.DB, pepper string, hmacSecretKey string) UserService {
	uGorm := &userGorm{db}
	return &userService{
		UserDB: newUserValidator(uGorm, pepper, hash.NewHMAC(hmacSecretKey)),
		pepper: pepper,
	}
}

type userService struct {
	UserDB
	pepper string
}

func (uService *userService) Authenticate(email, password string) (*User, error) {
	user, err := uService.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+uService.pepper))
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrIncorrectPassword
	case nil:
		return user, nil
	default:
		return nil, err
	}
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

func newUserValidator(udb UserDB, pepper string, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
		pepper:     pepper,
	}
}

func (uValidator *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user,
		uValidator.normalizeEmail); err != nil {
		return nil, err
	}
	return uValidator.UserDB.ByEmail((user.Email))
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
		uValidator.requirePassword,
		uValidator.passwordLength,
		uValidator.hashPassword,
		uValidator.requirePasswordHash,
		uValidator.generateRemember,
		uValidator.rememberLength,
		uValidator.hashRemember,
		uValidator.requireRememberHash,
		uValidator.requireEmail,
		uValidator.emailFormat,
		uValidator.normalizeEmail,
		uValidator.duplicateEmail); err != nil {
		return err
	}

	return uValidator.UserDB.Create(user)
}

func (uValidator *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user,
		uValidator.passwordLength,
		uValidator.hashPassword,
		uValidator.requirePasswordHash,
		uValidator.rememberLength,
		uValidator.hashRemember,
		uValidator.requireRememberHash,
		uValidator.requireEmail,
		uValidator.emailFormat,
		uValidator.normalizeEmail,
		uValidator.duplicateEmail); err != nil {
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
	pwBytes := []byte(user.Password + uValidator.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uValidator *userValidator) passwordLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func (uValidator *userValidator) requirePassword(user *User) error {
	if user.Password == "" {
		return ErrRequiredPassword
	}
	return nil
}

func (uValidator *userValidator) requirePasswordHash(user *User) error {
	if user.PasswordHash == "" {
		return ErrRequiredPassword
	}
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

func (uValidator *userValidator) rememberLength(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.Nbytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrInvalidRemeber
	}
	return nil
}

func (uValidator *userValidator) requireRememberHash(user *User) error {
	if user.RememberHash == "" {
		return ErrRequiredRememberHash
	}
	return nil
}

func (uValidator *userValidator) validateID(user *User) error {
	if user.ID <= 0 {
		return ErrInvalidID
	}
	return nil
}

func (uValidator *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	return nil
}

func (uValidator *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrRequiredEmail
	}
	return nil
}

func (uValidator *userValidator) emailFormat(user *User) error {
	if !uValidator.emailRegex.MatchString(user.Email) {
		return ErrInvalidEmail
	}
	return nil
}

func (uValidator *userValidator) duplicateEmail(user *User) error {
	existingUser, err := uValidator.ByEmail(user.Email)
	if err != nil {
		if err == ErrNotFound {
			return nil
		} else {
			return err
		}
	}
	if existingUser.ID != user.ID {
		return ErrTakenEmail
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
	//Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Methods for querying single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(RememberToken string) (*User, error)
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
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
