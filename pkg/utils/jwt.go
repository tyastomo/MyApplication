package utils

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTCustomClaims defines the custom claims for the JWT
type JWTCustomClaims struct {
	UserID   uuid.UUID `json:"userID"`
	UserType string    `json:"userType"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token
func GenerateJWT(userID uuid.UUID, userType string, jwtSecret string) (string, error) {
	claims := JWTCustomClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "payslip-generator", // Optional: Issuer
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

// ParseJWT validates and parses a JWT token string
func ParseJWT(tokenString string, jwtSecret string) (*jwt.Token, *JWTCustomClaims, error) {
	claims := &JWTCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	return token, claims, nil
}

// GetUserIDFromContext extracts the userID from Fiber context locals
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return uuid.Nil, fmt.Errorf("userID not found in context")
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("userID in context is not of type uuid.UUID")
	}
	return userID, nil
}

// GetUserTypeFromContext extracts the userType from Fiber context locals
func GetUserTypeFromContext(c *fiber.Ctx) (string, error) {
	userTypeVal := c.Locals("userType")
	if userTypeVal == nil {
		return "", fmt.Errorf("userType not found in context")
	}
	userType, ok := userTypeVal.(string)
	if !ok {
		return "", fmt.Errorf("userType in context is not of type string")
	}
	return userType, nil
}
