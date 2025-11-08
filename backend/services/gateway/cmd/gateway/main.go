package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/clients"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/config"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/handlers"
	"github.com/dongtandung2001/Doc-Agent/backend/services/gateway/internal/http"
)

func main() {
	// Load configuration (removed unused parameter)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize gRPC clients to all backend services
	// localAgentClient, localAgentConn, err := clients.NewLocalAgentClient(
	// 	cfg.Backends.LocalAgentService.Host,
	// 	cfg.Backends.LocalAgentService.Port,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to local agent service: %v", err)
	// }
	// defer localAgentConn.Close()

	codebaseClient, codebaseConn, err := clients.NewCodebaseAnalysisClient(
		cfg.Backends.CodebaseService.Host,
		cfg.Backends.CodebaseService.Port,
	)
	if err != nil {
		log.Fatalf("Failed t	o connect to codebase analysis service: %v", err)
	}
	defer codebaseConn.Close()

	// databaseClient, databaseConn, err := clients.NewDatabaseClient(
	// 	"localhost", // TODO: Get from config
	// 	9002,        // TODO: Get from config
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to database service: %v", err)
	// }
	// defer databaseConn.Close()

	// aiClient, aiConn, err := clients.NewAIClient(
	// 	cfg.Backends.AIService.Host,
	// 	cfg.Backends.AIService.Port,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to AI service: %v", err)
	// }
	// defer aiConn.Close()

	// Initialize gateway handler (proxies to all backend services)
	gatewayHandler := handlers.NewGatewayHandler(
		nil,
		codebaseClient,
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
