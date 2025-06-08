package utils

import (
	"os"
	"payslip-generator/pkg/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testJWTSecret = "test-secret-for-jwt"

func TestMain(m *testing.M) {
	// Setup: Use a fixed secret for tests to ensure reproducibility.
	// The config.LoadConfig() might be called by other test packages,
	// so explicitly set AppConfig for JWT tests if needed, or ensure .env.test is loaded.
	// For JWT tests specifically, we can override the secret if LoadConfig isn't consistently setting it first.
	os.Setenv("APP_ENV", "test") // Ensure test config path might be chosen if config.LoadConfig is called
	config.LoadConfig()          // Load .env.test which should set JWT_SECRET

	if config.AppConfig.JWTSecret == "" {
		// Fallback if .env.test wasn't loaded or JWT_SECRET was missing
		config.AppConfig.JWTSecret = testJWTSecret
	}
	// Ensure the secret used in tests is the one from AppConfig
	testJWTSecret = config.AppConfig.JWTSecret
	if testJWTSecret == "" {
		panic("JWT_SECRET for testing is empty")
	}


	exitCode := m.Run()
	os.Exit(exitCode)
}


func TestGenerateAndParseJWT(t *testing.T) {
	userID := uuid.New()
	userType := "employee"

	tokenString, err := GenerateJWT(userID, userType, testJWTSecret)
	require.NoError(t, err, "Should generate token without error")
	require.NotEmpty(t, tokenString, "Generated token string should not be empty")

	parsedToken, claims, err := ParseJWT(tokenString, testJWTSecret)
	require.NoError(t, err, "Should parse valid token without error")
	require.NotNil(t, parsedToken, "Parsed token should not be nil")
	require.True(t, parsedToken.Valid, "Parsed token should be valid")

	assert.Equal(t, userID, claims.UserID, "UserID in claims should match original")
	assert.Equal(t, userType, claims.UserType, "UserType in claims should match original")
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 5*time.Second, "Expiration time should be approximately 24 hours from now")
	assert.Equal(t, "payslip-generator", claims.Issuer, "Issuer should be as set")
}

func TestParseJWT_InvalidToken(t *testing.T) {
	// Test with a completely malformed token
	_, _, err := ParseJWT("this.is.not.a.jwt", testJWTSecret)
	assert.Error(t, err, "Should return error for malformed token")

	// Test with a token signed with a different secret
	userID := uuid.New()
	userType := "admin"
	wrongSecret := "another-secret-key"
	tokenSignedWithWrongSecret, _ := GenerateJWT(userID, userType, wrongSecret)

	_, _, err = ParseJWT(tokenSignedWithWrongSecret, testJWTSecret)
	assert.Error(t, err, "Should return error for token signed with wrong secret")
	// The error from jwt-go v5 for signature mismatch is "signature is invalid"
	// For an expired token, it's "token has invalid claims: token is expired"
	// For a token used before NBF, it's "token has invalid claims: token is not valid yet"
	// The wrapper in ParseJWT returns "failed to parse token: ..." or "invalid token"
	// We can check for specific wrapped errors if needed, but for now, just Error is fine.
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	userType := "employee"

	// Create a token that expired 1 hour ago
	claims := JWTCustomClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "payslip-generator",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredTokenString, _ := token.SignedString([]byte(testJWTSecret))

	_, _, err := ParseJWT(expiredTokenString, testJWTSecret)
	assert.Error(t, err, "Should return error for expired token")
	// Depending on how ParseJWT wraps errors, you might check for a specific error type or message
	// e.g. assert.Contains(t, err.Error(), jwt.ErrTokenExpired.Error()) or "token is expired"
}

func TestParseJWT_NotYetValidToken(t *testing.T) {
	userID := uuid.New()
	userType := "employee"

	claims := JWTCustomClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // Not valid for 30 minutes
			Issuer:    "payslip-generator",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	notYetValidTokenString, _ := token.SignedString([]byte(testJWTSecret))

	_, _, err := ParseJWT(notYetValidTokenString, testJWTSecret)
	assert.Error(t, err, "Should return error for token that is not yet valid (NBF)")
}

func TestGenerateJWT_EmptySecret(t *testing.T) {
	userID := uuid.New()
	userType := "employee"
	_, err := GenerateJWT(userID, userType, "")
	assert.Error(t, err, "Should return error if JWT secret is empty")
	assert.Contains(t, err.Error(), "key is of invalid type", "Error message should indicate key issue for empty secret")
}
