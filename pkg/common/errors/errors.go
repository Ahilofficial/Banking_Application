package errors

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// Standard Banking Error Codes
const (
	ErrCodeInvalidRequest      = "INVALID_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeInternal            = "INTERNAL_SERVER_ERROR"
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
	ErrCodeAccountBlocked      = "ACCOUNT_BLOCKED"
	ErrCodeAccountClosed       = "ACCOUNT_CLOSED"
	ErrCodeAccountNotFound     = "ACCOUNT_NOT_FOUND"
	ErrCodeCustomerNotFound    = "CUSTOMER_NOT_FOUND"
	ErrCodeUserAlreadyExists   = "USER_ALREADY_EXISTS"
	ErrCodeInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired      = "SESSION_EXPIRED"
	ErrCodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"
	ErrCodeTransferFailed      = "TRANSFER_FAILED"
	ErrCodeSameAccountTransfer = "SAME_ACCOUNT_TRANSFER"
)

// AppError represents a structured domain error for banking operations
type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(status int, code string, message string, detail ...string) *AppError {
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    message,
		Detail:     d,
	}
}

// Predefined error helpers
func ErrBadRequest(code, message string, detail ...string) *AppError {
	return NewAppError(http.StatusBadRequest, code, message, detail...)
}

func ErrUnauth(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

func ErrForbid(message string) *AppError {
	return NewAppError(http.StatusForbidden, ErrCodeForbidden, message)
}

func ErrResourceNotFound(code, message string) *AppError {
	return NewAppError(http.StatusNotFound, code, message)
}

func ErrInternalServerError(message string) *AppError {
	return NewAppError(http.StatusInternalServerError, ErrCodeInternal, message)
}

// FiberErrorHandler provides centralized error handling for Fiber v3
func FiberErrorHandler(c fiber.Ctx, err error) error {
	reqID := c.Get("X-Request-ID")
	if reqID == "" {
		if id, ok := c.Locals("requestid").(string); ok {
			reqID = id
		}
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
				"detail":  appErr.Detail,
			},
			"request_id": reqID,
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    http.StatusText(fiberErr.Code),
				"message": fiberErr.Message,
			},
			"request_id": reqID,
		})
	}

	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    ErrCodeInternal,
			"message": "An unexpected error occurred. Please try again later.",
		},
		"request_id": reqID,
	})
}
