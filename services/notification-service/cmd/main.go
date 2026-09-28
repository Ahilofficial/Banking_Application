package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"banking-microservices/pkg/common/config"
	"banking-microservices/services/notification-service/di"
)

func main() {
	cfg := config.LoadConfig("notification-service", 8006)

	bundle, err := di.InitializeServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize notification-service: %v", err)
	}

	// 1. Start gRPC Server
	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %d: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("[notification-service] gRPC Server listening on port %d", cfg.GRPCPort)
		if err := bundle.GRPCServer.Serve(lis); err != nil {
			log.Printf("[notification-service] gRPC server exited: %v", err)
		}
	}()

	// 2. Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n[notification-service] Shutting down gRPC and HTTP servers...")
		bundle.GRPCServer.GracefulStop()
		_ = bundle.App.Shutdown()
	}()

	// 3. Start Fiber HTTP REST Server
	httpAddr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[notification-service] Fiber HTTP Server listening on port %d", cfg.Port)
	if err := bundle.App.Listen(httpAddr); err != nil {
		log.Printf("[notification-service] HTTP server shutdown: %v", err)
	}
}
