package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/notification-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all notification service HTTP endpoints
func SetupRoutes(app *fiber.App, h *handlers.NotificationHandler, jwtSecret string) {
	api := app.Group("/api/v1/notifications")

	// Protected routes
	api.Use(middleware.AuthRequired(jwtSecret))

	api.Post("/email", h.SendEmail)
	api.Post("/sms", h.SendSMS)
	api.Post("/send", h.SendNotification)
	api.Get("", h.List)
	api.Get("/:id", h.GetByID)
}
