package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/gateway/di"
)

func main() {
	cfg := config.LoadConfig("api-gateway", 8000)

	app, err := di.InitializeGateway(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize API gateway: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[api-gateway] Fiber v3 API Gateway listening on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Gateway server terminated: %v", err)
	}
}
