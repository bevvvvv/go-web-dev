package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

func NewHMAC(key string) HMAC {
	hmac := hmac.New(sha256.New, []byte(key))
	return HMAC{
		hmac: hmac,
	}
}

type HMAC struct {
	hmac hash.Hash
}

func (hmac HMAC) Hash(input string) string {
	hmac.hmac.Reset()
	hmac.hmac.Write([]byte(input))
	bytes := hmac.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(bytes)
}
