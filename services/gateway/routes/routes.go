package routes

import (
	"fmt"
	"strings"
	"time"

	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/middleware"
	"banking-microservices/pkg/common/security"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
)

// SetupRoutes registers all API Gateway proxy routes and health checks
func SetupRoutes(app *fiber.App, cfg *config.Config) {
	// Health Check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "healthy",
			"gateway":   "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Forwarder helper
	forwardTo := func(baseURL string) fiber.Handler {
		return func(c fiber.Ctx) error {
			// Propagate user identity headers downstream if authenticated
			if userID := security.GetUserID(c); userID > 0 {
				c.Request().Header.Set(security.HeaderUserID, fmt.Sprintf("%d", userID))
			}
			if email := security.GetUserEmail(c); email != "" {
				c.Request().Header.Set(security.HeaderUserEmail, email)
			}
			if roles := security.GetUserRoles(c); len(roles) > 0 {
				c.Request().Header.Set(security.HeaderUserRoles, strings.Join(roles, ","))
			}
			if perms := security.GetUserPermissions(c); len(perms) > 0 {
				c.Request().Header.Set(security.HeaderUserPermissions, strings.Join(perms, ","))
			}
			if idemp := c.Get(security.HeaderIdempotencyKey); idemp != "" {
				c.Request().Header.Set(security.HeaderIdempotencyKey, idemp)
			}
			if reqID := security.GetRequestID(c); reqID != "" {
				c.Request().Header.Set(security.HeaderRequestID, reqID)
			}

			targetURL := baseURL + c.OriginalURL()
			return proxy.Do(c, targetURL)
		}
	}

	api := app.Group("/api/v1")

	// 1. Auth Service Routes
	// Public auth routes (register, login, refresh)
	api.Post("/auth/register", forwardTo(cfg.AuthServiceURL))
	api.Post("/auth/login", forwardTo(cfg.AuthServiceURL))
	api.Post("/auth/refresh", forwardTo(cfg.AuthServiceURL))
	api.Post("/auth/logout", forwardTo(cfg.AuthServiceURL))

	// Protected auth routes
	authProtected := api.Group("/auth", middleware.AuthRequired(cfg.JWTSecret))
	authProtected.Get("/me", forwardTo(cfg.AuthServiceURL))
	authProtected.Get("/users", middleware.RequirePermission("user:manage"), forwardTo(cfg.AuthServiceURL))
	authProtected.Get("/users/:id", forwardTo(cfg.AuthServiceURL))
	authProtected.Post("/roles/assign", middleware.RequirePermission("role:manage"), forwardTo(cfg.AuthServiceURL))

	// 2. Customer Service Routes (Protected)
	customerGroup := api.Group("/customers", middleware.AuthRequired(cfg.JWTSecret))
	customerGroup.Post("", forwardTo(cfg.CustomerServiceURL))
	customerGroup.Get("", middleware.RequirePermission("customer:view"), forwardTo(cfg.CustomerServiceURL))
	customerGroup.Get("/user/:user_id", forwardTo(cfg.CustomerServiceURL))
	customerGroup.Get("/:id", forwardTo(cfg.CustomerServiceURL))
	customerGroup.Patch("/:id/kyc", middleware.RequirePermission("customer:update"), forwardTo(cfg.CustomerServiceURL))

	// 3. Account Service Routes (Protected)
	accountGroup := api.Group("/accounts", middleware.AuthRequired(cfg.JWTSecret))
	accountGroup.Post("", middleware.RequirePermission("account:create"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Get("", middleware.RequirePermission("account:view"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Get("/number/:account_number", middleware.RequirePermission("account:view"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Get("/:id", middleware.RequirePermission("account:view"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Post("/:id/block", middleware.RequirePermission("account:block"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Post("/:id/unblock", middleware.RequirePermission("account:unblock"), forwardTo(cfg.AccountServiceURL))
	accountGroup.Post("/:id/close", middleware.RequirePermission("account:close"), forwardTo(cfg.AccountServiceURL))

	// 4. Transaction Service Routes (Protected)
	txGroup := api.Group("/transactions", middleware.AuthRequired(cfg.JWTSecret))
	txGroup.Post("/transfer", middleware.RequirePermission("transaction:create"), forwardTo(cfg.TransactionServiceURL))
	txGroup.Post("/deposit", middleware.RequirePermission("transaction:create"), forwardTo(cfg.TransactionServiceURL))
	txGroup.Post("/withdraw", middleware.RequirePermission("transaction:create"), forwardTo(cfg.TransactionServiceURL))
	txGroup.Get("/ledger", middleware.RequirePermission("ledger:view"), forwardTo(cfg.TransactionServiceURL))
	txGroup.Get("", middleware.RequirePermission("transaction:view"), forwardTo(cfg.TransactionServiceURL))
	txGroup.Get("/:id", middleware.RequirePermission("transaction:view"), forwardTo(cfg.TransactionServiceURL))

	// 5. Audit Service Routes (Protected)
	auditGroup := api.Group("/audit", middleware.AuthRequired(cfg.JWTSecret))
	auditGroup.Get("/logs", middleware.RequirePermission("audit:view"), forwardTo(cfg.AuditServiceURL))

	// 6. Notification Service Routes (Protected)
	notifGroup := api.Group("/notifications", middleware.AuthRequired(cfg.JWTSecret))
	notifGroup.Post("/email", forwardTo(cfg.NotificationServiceURL))
	notifGroup.Post("/sms", forwardTo(cfg.NotificationServiceURL))
	notifGroup.Post("/send", forwardTo(cfg.NotificationServiceURL))
	notifGroup.Get("", forwardTo(cfg.NotificationServiceURL))
	notifGroup.Get("/:id", forwardTo(cfg.NotificationServiceURL))
}
