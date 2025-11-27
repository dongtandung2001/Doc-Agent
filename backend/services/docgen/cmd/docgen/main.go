package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/config"
	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/service"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to AI Service
	aiClient, err := clients.NewAIClient(cfg.AI.Host, cfg.AI.Port)
	if err != nil {
		log.Fatalf("Failed to connect to AI Service: %v", err)
	}
	defer aiClient.Close()

	// Connect to Gateway (to reach Local Agent)
	gatewayClient, err := clients.NewGatewayClient(cfg.Gateway.Host, cfg.Gateway.Port)
	if err != nil {
		log.Fatalf("Failed to connect to Gateway: %v", err)
	}
	defer gatewayClient.Close()

	// Connect to Redis for task queue
	redisClient, err := clients.NewRedisClient(clients.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize service layer
	docgenSvc := service.NewDocGenService(aiClient, gatewayClient, redisClient)

	log.Println("🚀 Document Generation Worker starting...")
	log.Printf("   Connected to AI Service at %s:%d", cfg.AI.Host, cfg.AI.Port)
	log.Printf("   Connected to Gateway at %s:%d", cfg.Gateway.Host, cfg.Gateway.Port)
	log.Printf("   Connected to Redis at %s:%d", cfg.Redis.Host, cfg.Redis.Port)

	// Start message queue worker
	go func() {
		if err := docgenSvc.StartWorker(); err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down document generation worker...")
	docgenSvc.Shutdown()
}
