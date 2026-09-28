//go:build wireinject
// +build wireinject

package di

import (
	"banking-microservices/pkg/common/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

// GatewaySet bundles all providers required for Gateway
var GatewaySet = wire.NewSet(
	ProvideGatewayApp,
)

// InitializeGateway wires dependencies and initializes the API Gateway Fiber App
func InitializeGateway(cfg *config.Config) (*fiber.App, error) {
	wire.Build(GatewaySet)
	return nil, nil
}
