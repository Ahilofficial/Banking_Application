package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/account-service/di"
)

func main() {
	cfg := config.LoadConfig("account-service", 8003)

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize account-service: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[account-service] Running on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
