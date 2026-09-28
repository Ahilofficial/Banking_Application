package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/account-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all account service endpoints
func SetupRoutes(app *fiber.App, h *handlers.AccountHandler, jwtSecret string) {
	api := app.Group("/api/v1/accounts")
	api.Use(middleware.AuthRequired(jwtSecret))

	api.Post("", middleware.RequirePermission("account:create"), h.Create)
	api.Get("", middleware.RequirePermission("account:view"), h.List)
	api.Get("/number/:account_number", middleware.RequirePermission("account:view"), h.GetByAccountNumber)
	api.Get("/:id", middleware.RequirePermission("account:view"), h.GetByID)
	api.Post("/:id/block", middleware.RequirePermission("account:block"), h.Block)
	api.Post("/:id/unblock", middleware.RequirePermission("account:unblock"), h.Unblock)
	api.Post("/:id/close", middleware.RequirePermission("account:close"), h.Close)

	// Internal verification endpoint for transaction service
	api.Post("/internal/verify", h.InternalVerify)
}
