package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/config"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/handlers"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/http"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
)

func main() {
	// Load configuration (removed unused parameter)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize gRPC clients to all backend services
	localAgentClient, err := clients.NewLocalAgentClient(
		cfg.Backends.LocalAgentService.Host,
		cfg.Backends.LocalAgentService.Port,
	)
	if err != nil {
		log.Printf("Failed to connect to local agent service: %v", err)
	} else {
		log.Printf("Connected to Local Agent Service at %s:%d", cfg.Backends.LocalAgentService.Host, cfg.Backends.LocalAgentService.Port)
	}
	defer localAgentClient.Close()

	// codebaseClient, err := clients.NewCodebaseClient(
	// 	cfg.Backends.CodebaseService.Host,
	// 	cfg.Backends.CodebaseService.Port,
	// )
	// if err != nil {
	// 	log.Printf("Failed to connect to codebase analysis service: %v", err)
	// } else {
	// 	log.Printf("Connected to Codebase Analysis Service at %s:%d", cfg.Backends.CodebaseService.Host, cfg.Backends.CodebaseService.Port)
	// }
	// defer codebaseClient.Close()

	// databaseClient, err := clients.NewDatabaseClient(
	// 	"localhost", // TODO: Get from config
	// 	9002,        // TODO: Get from config
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to database service: %v", err)
	// }
	// defer databaseClient.Close()

	// aiClient, err := clients.NewAIClient(
	// 	cfg.Backends.AIService.Host,
	// 	cfg.Backends.AIService.Port,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to AI service: %v", err)
	// }
	// defer aiClient.Close()

	// Initialize gateway handler (proxies to all backend services)
	// Pass the underlying gRPC clients using GetClient()
	gatewayHandler := handlers.NewGatewayHandler(
		localAgentClient.GetClient(),
		nil,
		nil,
		nil,
	)

	// Create HTTP server with gateway handler
	srv := http.NewServer(cfg, gatewayHandler)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutdown signal received")
		srv.Shutdown()
	}()

	// Start server
	log.Println("Starting API Gateway...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Gateway failed to start: %v", err)
	}
}
