package models

import "strings"

var (
	ErrNotFound             modelError = "models: resource not found"
	ErrInvalidID            modelError = "models: ID provided is not valid"
	ErrInvalidEmail         modelError = "models: Email address is not valid"
	ErrRequiredEmail        modelError = "models: Email address is required"
	ErrTakenEmail           modelError = "models: Email address is already taken"
	ErrInvalidPassword      modelError = "models: Password is not valid"
	ErrRequiredPassword     modelError = "models: Password is required"
	ErrIncorrectPassword    modelError = "models: Incorrect password provided"
	ErrInvalidRemeber       modelError = "models: Remember token must be an adequate length"
	ErrRequiredRememberHash modelError = "models: Remember hash is required"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	str := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(str, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}
