package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/customer-service/di"
)

func main() {
	cfg := config.LoadConfig("customer-service", 8002)

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize customer-service: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[customer-service] Running on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
