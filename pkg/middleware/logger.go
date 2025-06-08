package middleware

import (
	"payslip-generator/pkg/constants"
	"payslip-generator/pkg/utils" // Your logger package
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Logger is a middleware that logs request details using Zap.
func Logger(c *fiber.Ctx) error {
	start := time.Now()

	// Process request
	err := c.Next()

	latency := time.Since(start)
	statusCode := c.Response().StatusCode()

	// Get values from context locals
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string) // Type assertion, default to empty if not string

	userIDVal := c.Locals(constants.UserIDKey.String())
	var userIDStr string
	if userID, ok := userIDVal.(uuid.UUID); ok {
		if userID != uuid.Nil {
			userIDStr = userID.String()
		}
	}

	userTypeVal := c.Locals(constants.UserTypeKey.String())
	userType, _ := userTypeVal.(string)

	// Prepare log fields
	fields := []zap.Field{
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Int("status_code", statusCode),
		zap.Duration("latency", latency),
		zap.String("ip_address", c.IP()),
		zap.String("user_agent", string(c.Request().Header.UserAgent())),
		zap.String("request_id", requestID),
	}

	if userIDStr != "" {
		fields = append(fields, zap.String("user_id", userIDStr))
	}
	if userType != "" {
		fields = append(fields, zap.String("user_type", userType))
	}

	// Log based on status code
	if statusCode >= 500 {
		utils.Logger.Error("Server error", fields...)
	} else if statusCode >= 400 {
		utils.Logger.Warn("Client error", fields...)
	} else {
		utils.Logger.Info("Request processed", fields...)
	}

	return err // Return error from c.Next() if any
}
