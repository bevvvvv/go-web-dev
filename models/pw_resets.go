package models

import (
	"go-web-dev/hash"
	"go-web-dev/rand"

	"github.com/jinzhu/gorm"
)

type pwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique_index"`
}

type pwResetValidator struct {
	pwResetDB
	hmac hash.HMAC
}

func newPWResetValidator(resetDB pwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		pwResetDB: resetDB,
		hmac:      hmac,
	}
}

func (resetValidator *pwResetValidator) Create(reset *pwReset) error {
	if err := runResetValFuncs(reset,
		resetValidator.requireUserID,
		resetValidator.generateToken,
		resetValidator.hashToken); err != nil {
		return err
	}

	return resetValidator.pwResetDB.Create(reset)
}

func (resetValidator *pwResetValidator) Delete(id uint) error {
	reset := pwReset{UserID: id}
	if err := runResetValFuncs(&reset,
		resetValidator.requireUserID); err != nil {
		return err
	}
	return resetValidator.pwResetDB.Delete(reset.UserID)
}

func (resetValidator *pwResetValidator) ByToken(token string) (*pwReset, error) {
	reset := pwReset{Token: token}
	if err := runResetValFuncs(&reset,
		resetValidator.hashToken); err != nil {
		return nil, err
	}

	return resetValidator.pwResetDB.ByToken(reset.TokenHash)
}

type resetValFunc func(*pwReset) error

func runResetValFuncs(reset *pwReset, funcs ...resetValFunc) error {
	for _, function := range funcs {
		if err := function(reset); err != nil {
			return err
		}
	}
	return nil
}

func (resetValidator *pwResetValidator) requireUserID(reset *pwReset) error {
	if reset.UserID <= 0 {
		return ErrRequiredUserID
	}
	return nil
}

func (resetValidator *pwResetValidator) hashToken(reset *pwReset) error {
	if reset.Token == "" {
		return nil
	}
	reset.TokenHash = resetValidator.hmac.Hash(reset.Token)
	return nil
}

func (resetValidator *pwResetValidator) generateToken(reset *pwReset) error {
	if reset.Token != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	reset.Token = token
	return nil
}

type pwResetDB interface {
	Create(reset *pwReset) error
	Delete(id uint) error

	ByToken(token string) (*pwReset, error)
}

type pwResetGorm struct {
	db *gorm.DB
}

func (resetGorm *pwResetGorm) Create(reset *pwReset) error {
	return resetGorm.db.Create(reset).Error
}

func (resetGorm *pwResetGorm) Delete(id uint) error {
	var reset pwReset
	reset.ID = id
	return resetGorm.db.Delete(&reset).Error
}

func (resetGorm *pwResetGorm) ByToken(tokenHash string) (*pwReset, error) {
	var reset pwReset
	err := first(resetGorm.db.Where("token_hash = ?", tokenHash), &reset)
	if err != nil {
		return nil, err
	}
	return &reset, nil
}
