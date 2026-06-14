package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GeneratePin returns a random 4-digit PIN as a zero-padded string.
func GeneratePin() (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate pin: %w", err)
	}
	n := (int(b[0])<<8 | int(b[1])) % 10000
	return fmt.Sprintf("%04d", n), nil
}

// HashPin bcrypt-hashes a PIN for storage.
func HashPin(pin string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash pin: %w", err)
	}
	return string(h), nil
}

// CheckPin reports whether pin matches the stored bcrypt hash.
func CheckPin(hash, pin string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil
}

// ConstantTimeEqual compares two secrets without leaking timing (admin PIN).
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidPin reports whether s is a 4-digit PIN.
func ValidPin(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
