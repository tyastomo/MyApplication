package middleware

import (
	"payslip-generator/pkg/constants"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID is a middleware that injects a unique request ID into the context and response header.
func RequestID(c *fiber.Ctx) error {
	// Generate a new UUID
	requestID := uuid.New().String()

	// Set the request ID in locals (for access within the application)
	c.Locals(constants.RequestIDKey.String(), requestID)

	// Set the request ID in the response header (for client-side access or tracing)
	c.Set("X-Request-ID", requestID)

	// Continue to the next middleware or handler
	return c.Next()
}
