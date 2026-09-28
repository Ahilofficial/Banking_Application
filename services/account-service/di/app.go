package di

import (
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/account-service/handlers"
	"banking-microservices/services/account-service/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// ProvideApp creates and configures the Fiber application with middlewares and routes
func ProvideApp(h *handlers.AccountHandler, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errors.FiberErrorHandler,
		AppName:      "Banking Microservices - Account Service",
	})

	app.Use(middleware.RequestID())
	app.Use(recover.New())

	// Register all routes from routes.go
	routes.SetupRoutes(app, h, cfg.JWTSecret)

	return app
}
