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
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/config"
	grpcserver "github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/grpc"
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize service layer
	dbSvc := service.NewDatabaseService()

	// Initialize gRPC server
	grpcSrv := grpc.NewServer()

	// Register DatabaseService
	dbServer := grpcserver.NewServer(dbSvc)
	apiv1.RegisterDatabaseServiceServer(grpcSrv, dbServer)

	// Start listening
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("🚀 Database Service listening on %s", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down database service...")
		grpcSrv.GracefulStop()
	}()

	// Start serving
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
