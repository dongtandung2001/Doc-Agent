package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/config"
	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/service"
)

func main() {
	// Load configuration
	_, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize service layer
	docgenSvc := service.NewDocGenService()

	log.Println("🚀 Document Generation Worker starting...")

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
}
