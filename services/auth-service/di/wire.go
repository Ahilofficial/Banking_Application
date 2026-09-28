//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	"banking-microservices/services/auth-service/handlers"
	"banking-microservices/services/auth-service/repository"
	"banking-microservices/services/auth-service/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// AuthSet bundles all providers required for Auth Service
var AuthSet = wire.NewSet(
	database.ProviderSet,
	client.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	handlers.NewAuthHandler,
	ProvideApp,
)

// InitializeApp wires dependencies and initializes the Fiber App
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(AuthSet)
	return nil, nil
}
