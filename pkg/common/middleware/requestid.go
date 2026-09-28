package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestID ensures every incoming request has an X-Request-ID header and local context
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("X-Request-ID", reqID)
		c.Locals("requestid", reqID)
		return c.Next()
	}
}
