//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/client"
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/database"
	grpcService "banking-microservices/services/notification-service/grpc"
	"banking-microservices/services/notification-service/handlers"
	"banking-microservices/services/notification-service/repository"
	"banking-microservices/services/notification-service/service"

	"github.com/google/wire"
)

// NotificationSet bundles all providers required for Notification Service
var NotificationSet = wire.NewSet(
	database.ProviderSet,
	client.ProviderSet,
	repository.ProviderSet,
	service.ProviderSet,
	grpcService.ProviderSet,
	handlers.ProviderSet,
	ProvideApp,
	ProvideGRPC,
	ProvideServerBundle,
)

// InitializeServer wires all dependencies and returns the complete server bundle
func InitializeServer(cfg *config.Config) (*ServerBundle, error) {
	wire.Build(NotificationSet)
	return nil, nil
}
