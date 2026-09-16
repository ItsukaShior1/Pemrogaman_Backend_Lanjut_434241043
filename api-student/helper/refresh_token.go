package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateRefreshToken() (raw, hashed string, err error) {
	b := make([]byte, 48)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = base64.RawURLEncoding.EncodeToString(sum[:])
	return raw, hashed, nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
