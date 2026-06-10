package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// apiKeyBytes is the entropy of a raw API key before encoding (256 bits).
const apiKeyBytes = 32

// GenerateApiKey returns a fresh CSPRNG API key (the plaintext to hand to the
// user exactly once) together with the hash to persist. The plaintext is never
// stored or logged.
func GenerateApiKey() (plaintext, hash string, err error) {
	buf := make([]byte, apiKeyBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = base64.RawURLEncoding.EncodeToString(buf)
	return plaintext, HashApiKey(plaintext), nil
}

// HashApiKey maps a raw API key to its at-rest representation. Lookups hash the
// incoming key and compare hashes, so the plaintext never needs to be stored.
func HashApiKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
