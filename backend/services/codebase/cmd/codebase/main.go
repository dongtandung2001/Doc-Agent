package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/config"
	grpcserver "github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/grpc"
	"github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/service"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("failed to create logs directory: %v", err)
	}
	file, err := os.OpenFile("logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)
	// Optionally, customize the log format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	aiClient, err := clients.NewAIClient("", 0)
	if err != nil {
		log.Fatalf("Failed to connect to AIService: %v", err)
	}
	defer aiClient.Close()

	gatewayClient, err := clients.NewGatewayClient(
		cfg.Gateway.Host,
		cfg.Gateway.Port,
	)
	if err != nil {
		log.Fatalf("Failed to connect to GatewayService: %v", err)
	}
	defer gatewayClient.Close()

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
	analysisSvc := service.NewAnalysisService(aiClient, gatewayClient, redisClient)

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
		clients.GetGlobalFileCache().Shutdown()
		grpcSrv.GracefulStop()
	}()

	// Start serving
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
