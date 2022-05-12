package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

func Bytes(nBytes int) ([]byte, error) {
	bytes := make([]byte, nBytes)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func String(nBytes int) (string, error) {
	bytes, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}
