package models

import "strings"

var (
	ErrNotFound             privateError = "models: resource not found"
	ErrInvalidID            privateError = "models: ID provided is not valid"
	ErrRequiredUserID       privateError = "models: ID is required"
	ErrInvalidEmail         modelError   = "models: Email address is not valid"
	ErrRequiredEmail        modelError   = "models: Email address is required"
	ErrTakenEmail           modelError   = "models: Email address is already taken"
	ErrInvalidPassword      modelError   = "models: Password is not valid"
	ErrRequiredPassword     modelError   = "models: Password is required"
	ErrIncorrectPassword    modelError   = "models: Incorrect password provided"
	ErrInvalidRemeber       privateError = "models: Remember token must be an adequate length"
	ErrRequiredRememberHash privateError = "models: Remember hash is required"
	ErrRequiredTitle        modelError   = "models: Title is required"
	ErrInvalidResetToken    modelError   = "models: Token provided is not valid"
	ErrExpiredResetToken    modelError   = "models: Token provided has expired"
	ErrRequiredServiceName  privateError = "models: Service name is required"
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

type privateError string

func (e privateError) Error() string {
	return string(e)
}
