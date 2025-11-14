package utils

import (
	"crypto/rand"
	"encoding/hex"
)

const APIKeyLength = 32

func GenerateAPIKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
