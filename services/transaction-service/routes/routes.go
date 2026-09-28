package routes

import (
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/transaction-service/handlers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all transaction and ledger endpoints
func SetupRoutes(app *fiber.App, h *handlers.TransactionHandler, jwtSecret string) {
	api := app.Group("/api/v1/transactions")
	api.Use(middleware.AuthRequired(jwtSecret))

	api.Post("/transfer", middleware.RequirePermission("transaction:create"), h.Transfer)
	api.Post("/deposit", middleware.RequirePermission("transaction:create"), h.Deposit)
	api.Post("/withdraw", middleware.RequirePermission("transaction:create"), h.Withdraw)
	api.Get("/ledger", middleware.RequirePermission("ledger:view"), h.ListLedger)
	api.Get("", middleware.RequirePermission("transaction:view"), h.List)
	api.Get("/:id", middleware.RequirePermission("transaction:view"), h.GetByID)
}
