package response

import "github.com/gofiber/fiber/v3"

// ApiResponse is the standard envelope for all API responses
type ApiResponse struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Success returns a standardized success JSON response
func Success(c fiber.Ctx, status int, data any, message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	reqID := c.Get("X-Request-ID")
	if reqID == "" {
		if id, ok := c.Locals("requestid").(string); ok {
			reqID = id
		}
	}

	return c.Status(status).JSON(ApiResponse{
		Success:   true,
		Data:      data,
		Message:   msg,
		RequestID: reqID,
	})
}
