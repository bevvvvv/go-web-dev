package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

const (
	OAuthDropbox = "dropbox"
)

type OAuth struct {
	gorm.Model
	UserID      uint   `gorm:"not null;unique_index:user_service"`
	ServiceName string `gorm:"not null;unique_index:user_service"`
	oauth2.Token
}

type OAuthService interface {
	OAuthDB
}

func NewOAuthService(db *gorm.DB) OAuthService {
	return &oAuthService{
		OAuthDB: &oAuthGorm{db},
	}
}

type oAuthService struct {
	OAuthDB
}

var _ OAuthDB = &oAuthValidator{}

type oAuthValidator struct {
	OAuthDB
}

func (oauthValidator *oAuthValidator) Create(oAuth *OAuth) error {
	err := runOAuthValFuncs(oAuth,
		oauthValidator.requireUserID,
		oauthValidator.requireServiceName)
	if err != nil {
		return err
	}
	return oauthValidator.OAuthDB.Create(oAuth)
}

func (oauthValidator *oAuthValidator) Delete(id uint) error {
	oAuth := OAuth{Model: gorm.Model{ID: id}}
	err := runOAuthValFuncs(&oAuth,
		oauthValidator.requireUserID,
		oauthValidator.requireServiceName)
	if err != nil {
		return err
	}
	return oauthValidator.OAuthDB.Delete(oAuth.ID)
}

func (oauthValidator *oAuthValidator) Find(userID uint, serviceName string) (*OAuth, error) {
	oAuth := OAuth{UserID: userID, ServiceName: serviceName}
	err := runOAuthValFuncs(&oAuth,
		oauthValidator.requireUserID,
		oauthValidator.requireServiceName)
	if err != nil {
		return nil, err
	}
	return oauthValidator.OAuthDB.Find(oAuth.UserID, oAuth.ServiceName)
}

type oAuthValFunc func(*OAuth) error

func runOAuthValFuncs(oAuth *OAuth, funcs ...oAuthValFunc) error {
	for _, function := range funcs {
		if err := function(oAuth); err != nil {
			return err
		}
	}
	return nil
}

func (oauthValidator *oAuthValidator) requireUserID(oAuth *OAuth) error {
	if oAuth.UserID <= 0 {
		return ErrRequiredUserID
	}
	return nil
}

func (oauthValidator *oAuthValidator) requireServiceName(oAuth *OAuth) error {
	if oAuth.ServiceName == "" {
		return ErrRequiredServiceName
	}
	return nil
}

type OAuthDB interface {
	Create(oauth *OAuth) error
	Delete(id uint) error

	Find(userID uint, serviceName string) (*OAuth, error)
}

var _ OAuthDB = &oAuthGorm{}

type oAuthGorm struct {
	db *gorm.DB
}

func (oauthGorm *oAuthGorm) Create(oAuth *OAuth) error {
	return oauthGorm.db.Create(oAuth).Error
}

func (oauthGorm *oAuthGorm) Delete(id uint) error {
	oAuth := OAuth{Model: gorm.Model{ID: id}}
	return oauthGorm.db.Unscoped().Delete(&oAuth).Error
}

func (oauthGorm *oAuthGorm) Find(userID uint, serviceName string) (*OAuth, error) {
	var oAuth OAuth
	db := oauthGorm.db.Where("user_id = ?", userID).Where("service_name = ?", serviceName)
	err := first(db, &oAuth)
	return &oAuth, err
}
