//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/services/account-service/handlers"
	"banking-microservices/services/account-service/repository"
	"banking-microservices/services/account-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// AccountSet bundles all providers required for Account Service
var AccountSet = wire.NewSet(
	database.ProviderSet,
	client.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	handlers.NewAccountHandler,
	ProvideApp,
)

// InitializeApp wires dependencies and initializes the Fiber App
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(AccountSet)
	return nil, nil
}
