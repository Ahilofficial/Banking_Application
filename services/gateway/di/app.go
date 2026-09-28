package di

import (
	"time"

	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/services/gateway/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// ProvideGatewayApp configures the API Gateway Fiber application
func ProvideGatewayApp(cfg *config.Config) *fiber.App {
	// Allow proxying to internal/loopback microservice addresses
	proxy.WithSecurityPolicy(proxy.SecurityPolicy{
		AllowPrivateIPs: true,
		AllowedSchemes:  []string{"http", "https"},
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: errors.FiberErrorHandler,
		AppName:      "Banking Microservices - API Gateway",
	})

	// Global Middleware Stack
	// 1. Request ID Tracing
	app.Use(middleware.RequestID())

	// 2. Structured Access Logger
	app.Use(logger.New())

	// 3. Panic Recovery
	app.Use(recover.New())

	// 4. CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "Idempotency-Key"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	// 5. Rate Limiting (120 req/min)
	app.Use(limiter.New(limiter.Config{
		Max:               120,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return errors.NewAppError(fiber.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded. Please wait before retrying.")
		},
	}))

	// Setup all gateway routes from routes.go
	routes.SetupRoutes(app, cfg)

	return app
}
