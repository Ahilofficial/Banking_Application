package security

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	HeaderUserID          = "X-User-ID"
	HeaderUserEmail       = "X-User-Email"
	HeaderUserRoles       = "X-User-Roles"
	HeaderUserPermissions = "X-User-Permissions"
	HeaderRequestID       = "X-Request-ID"
	HeaderIdempotencyKey  = "Idempotency-Key"

	LocalUserID          = "user_id"
	LocalUserEmail       = "user_email"
	LocalUserRoles       = "user_roles"
	LocalUserPermissions = "user_permissions"
)

// GetUserID extracts the user ID from Fiber locals or the X-User-ID header
func GetUserID(c fiber.Ctx) uint {
	if val := c.Locals(LocalUserID); val != nil {
		if id, ok := val.(uint); ok {
			return id
		}
		if id, ok := val.(int); ok {
			return uint(id)
		}
		if id, ok := val.(float64); ok {
			return uint(id)
		}
	}

	headerVal := c.Get(HeaderUserID)
	if headerVal != "" {
		if id, err := strconv.ParseUint(headerVal, 10, 64); err == nil {
			return uint(id)
		}
	}

	return 0
}

// GetUserEmail extracts the user's email
func GetUserEmail(c fiber.Ctx) string {
	if val := c.Locals(LocalUserEmail); val != nil {
		if email, ok := val.(string); ok {
			return email
		}
	}
	return c.Get(HeaderUserEmail)
}

// GetUserRoles extracts user roles
func GetUserRoles(c fiber.Ctx) []string {
	if val := c.Locals(LocalUserRoles); val != nil {
		if roles, ok := val.([]string); ok {
			return roles
		}
	}
	headerVal := c.Get(HeaderUserRoles)
	if headerVal != "" {
		return strings.Split(headerVal, ",")
	}
	return nil
}

// GetUserPermissions extracts user permissions
func GetUserPermissions(c fiber.Ctx) []string {
	if val := c.Locals(LocalUserPermissions); val != nil {
		if perms, ok := val.([]string); ok {
			return perms
		}
	}
	headerVal := c.Get(HeaderUserPermissions)
	if headerVal != "" {
		return strings.Split(headerVal, ",")
	}
	return nil
}

// HasPermission checks if the context has a given permission
func HasPermission(c fiber.Ctx, requiredPermission string) bool {
	// Super admin bypass
	roles := GetUserRoles(c)
	for _, r := range roles {
		if r == "SUPER_ADMIN" {
			return true
		}
	}

	perms := GetUserPermissions(c)
	for _, p := range perms {
		if p == requiredPermission || p == "*" {
			return true
		}
	}
	return false
}

// GetRequestID gets request ID from header or locals
func GetRequestID(c fiber.Ctx) string {
	reqID := c.Get(HeaderRequestID)
	if reqID != "" {
		return reqID
	}
	if id, ok := c.Locals("requestid").(string); ok {
		return id
	}
	return ""
}
