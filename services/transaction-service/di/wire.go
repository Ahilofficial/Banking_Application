//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/services/transaction-service/handlers"
	"banking-microservices/services/transaction-service/repository"
	"banking-microservices/services/transaction-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// TransactionSet bundles all providers required for Transaction Service
var TransactionSet = wire.NewSet(
	database.ProviderSet,
	client.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	handlers.NewTransactionHandler,
	ProvideApp,
)

// InitializeApp wires dependencies and initializes the Fiber App
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(TransactionSet)
	return nil, nil
}
