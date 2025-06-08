package middleware

import (
	"payslip-generator/pkg/config"
	"payslip-generator/pkg/constants"
	"payslip-generator/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// DeserializeUser is a middleware to authenticate users via JWT
func DeserializeUser(c *fiber.Ctx) error {
	var tokenString string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		tokenString = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("token") != "" { // Fallback to cookie if header not present
		tokenString = c.Cookies("token")
	}

	if tokenString == "" {
		return c.Next() // No token, proceed but user will not be authenticated
	}

	_, claims, err := utils.ParseJWT(tokenString, config.AppConfig.JWTSecret)
	if err != nil {
		// Differentiate between an expired token and other parsing errors for logging or specific responses
		if err == jwt.ErrTokenExpired {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Token has expired"})
		}
		// For other errors, you might not want to expose details
		// log.Printf("Error parsing token: %v\n", err) // Good for server logs
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid token"})
	}

	c.Locals(constants.UserIDKey.String(), claims.UserID)
	c.Locals(constants.UserTypeKey.String(), claims.UserType)
	// Example: c.Locals("user", claims) // if you want to store all claims

	return c.Next()
}

// RequireLoggedIn checks if a user is logged in (i.e., userID is in context)
func RequireLoggedIn() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals(constants.UserIDKey.String())
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "fail",
				"message": "You are not logged in. Please provide a valid token.",
			})
		}
		return c.Next()
	}
}

// RequireUserType creates a middleware to check for a specific user type.
func RequireUserType(requiredType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware should run after RequireLoggedIn or ensure userID is checked
		userID := c.Locals(constants.UserIDKey.String())
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "fail",
				"message": "Authentication required. Please log in.",
			})
		}

		userType, ok := c.Locals(constants.UserTypeKey.String()).(string)
		if !ok || userType == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "User type not found in token or token is invalid.",
			})
		}

		if userType != requiredType {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "You are not authorized to perform this action. Required role: " + requiredType,
			})
		}
		return c.Next()
	}
}
