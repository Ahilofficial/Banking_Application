package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/auth-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all authentication and RBAC endpoints
func SetupRoutes(app *fiber.App, h *handlers.AuthHandler, jwtSecret string) {
	api := app.Group("/api/v1/auth")

	// Public routes
	api.Post("/register", h.Register)
	api.Post("/login", h.Login)
	api.Post("/refresh", h.Refresh)
	api.Post("/logout", h.Logout)
	api.Post("/introspect", h.Introspect)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(jwtSecret))
	protected.Get("/me", h.Me)
	protected.Get("/users", middleware.RequirePermission("user:manage"), h.ListUsers)
	protected.Get("/users/:id", h.GetUserByID)
	protected.Post("/roles/assign", middleware.RequirePermission("role:manage"), h.AssignRole)
}
