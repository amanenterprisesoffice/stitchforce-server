package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"regexp"

	"github.com/google/uuid"
)

func NewID() (string, error) {
	return uuid.NewString(), nil
}

// IsValidEmail makes a basic validation of the email format using regex.
func IsValidEmail(email string) bool {
	// Simple regex for email validation
	re := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(re, email)
	return matched
}

// GenerateToken generates a cryptographically secure random token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
