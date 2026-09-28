package di

import (
	"banking-microservices/pkg/common/config"
	"banking-microservices/pkg/common/errors"
	"banking-microservices/pkg/common/middleware"
	grpcService "banking-microservices/services/notification-service/grpc"
	"banking-microservices/services/notification-service/handlers"
	"banking-microservices/services/notification-service/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"google.golang.org/grpc"
)

// ServerBundle bundles both Fiber HTTP App and gRPC Server
type ServerBundle struct {
	App        *fiber.App
	GRPCServer *grpc.Server
}

// ProvideApp creates and configures the Fiber application with middlewares and routes
func ProvideApp(h *handlers.NotificationHandler, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errors.FiberErrorHandler,
		AppName:      "Banking Microservices - Notification Service",
	})

	app.Use(middleware.RequestID())
	app.Use(recover.New())

	// Register all HTTP routes from routes.go
	routes.SetupRoutes(app, h, cfg.JWTSecret)

	return app
}

// ProvideGRPC initializes and registers the gRPC server
func ProvideGRPC(grpcImpl *grpcService.NotificationGRPCServer) *grpc.Server {
	s := grpc.NewServer()
	grpcImpl.Register(s)
	return s
}

// ProvideServerBundle wraps both HTTP and gRPC servers
func ProvideServerBundle(app *fiber.App, grpcServer *grpc.Server) *ServerBundle {
	return &ServerBundle{
		App:        app,
		GRPCServer: grpcServer,
	}
}
