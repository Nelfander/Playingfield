package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain text password and returns a bcrypt hash.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// wrap the error so i know it happened during hashing
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPasswordHash compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns true if it's a match.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
