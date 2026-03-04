package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/config"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"

	// "github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/db"
	grpcserver "github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/grpc"
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/repository"
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Database.PostgresURL == "" {
		log.Fatal("PostgreSQL URL is required (set POSTGRES_URL or database.postgres_url)")
	}

	// // Run migrations
	// if err := db.RunMigrations(cfg.Database.PostgresURL); err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }
	// log.Println("Database migrations completed")

	// Connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize repositories and service layer
	sectionRepo := repository.NewPostgresSectionRepository(pool)
	fileRepo := repository.NewPostgresFileItemRepository(pool)
	dbSvc := service.NewDatabaseService(sectionRepo, fileRepo)

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
