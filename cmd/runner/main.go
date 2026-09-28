package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"banking-microservices/pkg/common/config"
)

type ServiceDef struct {
	Name string
	Path string
	Port int
}

func main() {
	cfg := config.LoadConfig("runner", 0)

	services := []ServiceDef{
		{Name: "audit-service", Path: "./services/audit-service/cmd/main.go", Port: 8005},
		{Name: "notification-service", Path: "./services/notification-service/cmd/main.go", Port: 8006},
		{Name: "auth-service", Path: "./services/auth-service/cmd/main.go", Port: 8001},
		{Name: "customer-service", Path: "./services/customer-service/cmd/main.go", Port: 8002},
		{Name: "account-service", Path: "./services/account-service/cmd/main.go", Port: 8003},
		{Name: "transaction-service", Path: "./services/transaction-service/cmd/main.go", Port: 8004},
		{Name: "api-gateway", Path: "./services/gateway/main.go", Port: 8000},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nShutdown signal received. Stopping all microservices...")
		cancel()
	}()

	var wg sync.WaitGroup

	for _, s := range services {
		wg.Add(1)
		go func(svc ServiceDef) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				log.Printf("Starting [%s] on port %d...", svc.Name, svc.Port)
				cmd := exec.CommandContext(ctx, "go", "run", svc.Path)
				cmd.Env = append(os.Environ(),
					fmt.Sprintf("PORT=%d", svc.Port),
					fmt.Sprintf("GRPC_PORT=%d", cfg.GRPCPort),
					fmt.Sprintf("DB_DRIVER=%s", cfg.DBDriver),
					fmt.Sprintf("DB_DSN=%s", cfg.DBDSN),
					fmt.Sprintf("JWT_SECRET=%s", cfg.JWTSecret),
					"AUTH_SERVICE_URL=http://localhost:8001",
					"CUSTOMER_SERVICE_URL=http://localhost:8002",
					"ACCOUNT_SERVICE_URL=http://localhost:8003",
					"TRANSACTION_SERVICE_URL=http://localhost:8004",
					"AUDIT_SERVICE_URL=http://localhost:8005",
					"NOTIFICATION_SERVICE_URL=http://localhost:8006",
					"NOTIFICATION_GRPC_ADDR=localhost:50051",
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil && ctx.Err() == nil {
					log.Printf("[%s] exited with error: %v. Restarting in 2s...", svc.Name, err)
					time.Sleep(2 * time.Second)
				} else {
					return
				}
			}
		}(s)
		time.Sleep(300 * time.Millisecond) // staggered start
	}

	log.Println("\n=======================================================")
	log.Println("Banking Microservices Architecture (Go + Fiber v3 + MySQL)")
	log.Printf("Database:            %s (%s)\n", cfg.DBDriver, cfg.DBDSN)
	log.Println("API Gateway:         http://localhost:8000")
	log.Println("Auth Service:        http://localhost:8001")
	log.Println("Customer Service:    http://localhost:8002")
	log.Println("Account Service:     http://localhost:8003")
	log.Println("Transaction Service: http://localhost:8004")
	log.Println("Audit Service:       http://localhost:8005")
	log.Println("Notification Service:http://localhost:8006 (gRPC: 50051)")
	log.Println("=======================================================\n")

	wg.Wait()
	log.Println("All microservices stopped.")
}
