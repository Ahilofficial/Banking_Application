//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/services/audit-service/handlers"
	"banking-microservices/services/audit-service/repository"
	"banking-microservices/services/audit-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// AuditSet bundles all providers required for Audit Service
var AuditSet = wire.NewSet(
	database.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	handlers.NewAuditHandler,
	ProvideApp,
)

// InitializeApp wires dependencies and initializes the Fiber App
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(AuditSet)
	return nil, nil
}
