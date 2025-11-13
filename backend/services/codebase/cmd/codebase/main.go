package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/config"
	grpcserver "github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/grpc"
	"github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/service"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	aiClient, err := clients.NewAIClient(cfg.AI.Host, cfg.AI.Port)
	if err != nil {
		log.Fatalf("Failed to connect to AIService: %v", err)
	}
	defer aiClient.Close()

	// Initialize service layer
	analysisSvc := service.NewAnalysisService(aiClient)

	// Initialize gRPC server
	grpcSrv := grpc.NewServer(
	// TODO: Add interceptors (logging, auth, etc.)
	)

	// Register CodebaseAnalysisService
	codebaseServer := grpcserver.NewServer(analysisSvc)
	apiv1.RegisterCodebaseAnalysisServiceServer(grpcSrv, codebaseServer)

	// Start listening
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("🚀 Codebase Analysis Service listening on %s", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down codebase analysis service...")
		grpcSrv.GracefulStop()
	}()

	// Start serving
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
