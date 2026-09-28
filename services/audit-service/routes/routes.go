package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/audit-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all audit endpoints
func SetupRoutes(app *fiber.App, h *handlers.AuditHandler, jwtSecret string) {
	api := app.Group("/api/v1/audit")

	// Endpoint to ingest audit logs (internal microservices)
	api.Post("/logs", h.Record)

	// Endpoint to read audit trail (requires audit:view)
	api.Get("/logs", middleware.AuthRequired(jwtSecret), middleware.RequirePermission("audit:view"), h.List)
}
