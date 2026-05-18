package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plain-text password using bcrypt.
// bcrypt automatically generates a salt and embeds it in the hash.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a plain-text password against a bcrypt hash.
// Returns nil if they match.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
