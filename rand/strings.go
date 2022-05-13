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

// Returns the number of bytes in the given base64 URL encoded string
func Nbytes(base64string string) (int, error) {
	bytes, err := base64.URLEncoding.DecodeString(base64string)
	if err != nil {
		return -1, err
	}
	return len(bytes), nil
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
