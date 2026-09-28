package main

import (
	"fmt"
	"log"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/audit-service/di"
)

func main() {
	cfg := config.LoadConfig("audit-service", 8005)

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize audit-service: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[audit-service] Running on port %d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
