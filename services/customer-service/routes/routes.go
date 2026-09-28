package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/customer-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all customer service endpoints
func SetupRoutes(app *fiber.App, h *handlers.CustomerHandler, jwtSecret string) {
	api := app.Group("/api/v1/customers")
	api.Use(middleware.AuthRequired(jwtSecret))

	api.Post("", h.Create)
	api.Get("", middleware.RequirePermission("customer:view"), h.List)
	api.Get("/user/:user_id", h.GetByUserID)
	api.Get("/:id", h.GetByID)
	api.Patch("/:id/kyc", middleware.RequirePermission("customer:update"), h.UpdateKYC)
}
