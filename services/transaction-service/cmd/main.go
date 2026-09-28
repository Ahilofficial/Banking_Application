package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/transaction-service/di"
)

func main() {
	cfg := config.LoadConfig("transaction-service", 8004)

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize transaction-service: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[transaction-service] Running on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
