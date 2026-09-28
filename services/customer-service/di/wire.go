//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/services/customer-service/handlers"
	"banking-microservices/services/customer-service/repository"
	"banking-microservices/services/customer-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// CustomerSet bundles all providers required for Customer Service
var CustomerSet = wire.NewSet(
	database.ProviderSet,
	client.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	handlers.NewCustomerHandler,
	ProvideApp,
)

// InitializeApp wires dependencies and initializes the Fiber App
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(CustomerSet)
	return nil, nil
}
