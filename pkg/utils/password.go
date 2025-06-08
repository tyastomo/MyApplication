package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain password and returns the bcrypt hash.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // DefaultCost is 10
	return string(bytes), err
}

// CheckPasswordHash compares a plain password with a hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
