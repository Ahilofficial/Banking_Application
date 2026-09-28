package middleware

import (
	"strings"

	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/security"

	"github.com/gofiber/fiber/v3"
)

// AuthRequired ensures request is authenticated via Bearer token
func AuthRequired(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return errors.ErrUnauth("Missing Authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return errors.ErrUnauth("Invalid Authorization header format. Expected 'Bearer <token>'")
		}

		tokenString := parts[1]
		claims, err := security.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return errors.ErrUnauth("Invalid or expired access token")
		}

		// Store identity in locals
		c.Locals(security.LocalUserID, claims.UserID)
		c.Locals(security.LocalUserEmail, claims.Email)
		c.Locals(security.LocalUserRoles, claims.Roles)
		c.Locals(security.LocalUserPermissions, claims.Permissions)

		return c.Next()
	}
}

// RequirePermission enforces Level 1 RBAC authorization
func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !security.HasPermission(c, permission) {
			return errors.ErrForbid("Forbidden: missing required permission '" + permission + "'")
		}
		return c.Next()
	}
}

// RequireAnyPermission checks if caller has at least one of the permissions
func RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		for _, perm := range permissions {
			if security.HasPermission(c, perm) {
				return c.Next()
			}
		}
		return errors.ErrForbid("Forbidden: missing required permission")
	}
}
