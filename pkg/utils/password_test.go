package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "plainpassword"
	hashedPassword, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Check that the hash is a valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err, "Hashed password should be verifiable with original password")
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mypassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Correct password
	assert.True(t, CheckPasswordHash(password, string(hashedPassword)), "Correct password should match hash")

	// Incorrect password
	assert.False(t, CheckPasswordHash("wrongpassword", string(hashedPassword)), "Incorrect password should not match hash")

	// Invalid hash (not a bcrypt hash)
	assert.False(t, CheckPasswordHash(password, "notarealhash"), "Valid password should not match an invalid hash format")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	password := ""
	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err, "Hashed empty password should be verifiable")
	assert.True(t, CheckPasswordHash(password, hashedPassword), "CheckPasswordHash should verify empty password correctly")
}
