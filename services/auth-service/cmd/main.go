package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/auth-service/di"
)

func main() {
	cfg := config.LoadConfig("auth-service", 8001)

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize auth-service: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[auth-service] Running on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
