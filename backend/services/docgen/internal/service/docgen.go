package service

import (
	"log"

	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/tasks"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	"github.com/hibiken/asynq"
)

type DocGenService struct {
	aiClient      *clients.AIClient
	gatewayClient *clients.GatewayClient
	redisClient   *clients.RedisClient
	dbClient      *clients.DatabaseClient
	server        *asynq.Server
}

func NewDocGenService(
	aiClient *clients.AIClient,
	gatewayClient *clients.GatewayClient,
	redisClient *clients.RedisClient,
	dbClient *clients.DatabaseClient,
) *DocGenService {
	// Create asynq server configuration
	srv := asynq.NewServer(
		redisClient.GetRedisOpt(),
		asynq.Config{
			Concurrency: 5, // Process up to 5 tasks concurrently
			Queues: map[string]int{
				"default": 1, // Priority level for default queue
			},
		},
	)

	return &DocGenService{
		aiClient:      aiClient,
		gatewayClient: gatewayClient,
		redisClient:   redisClient,
		dbClient:      dbClient,
		server:        srv,
	}
}

// StartWorker starts the asynq server to consume tasks from Redis MQ
func (s *DocGenService) StartWorker() error {
	log.Println("Document generation worker started")
	log.Println("Waiting for tasks from Redis message queue...")

	// Create task handler with dependencies
	taskHandler := tasks.NewTaskHandler(s.aiClient, s.gatewayClient, s.dbClient)

	// Register all task handlers
	mux := taskHandler.RegisterHandlers()

	// Start the asynq server (blocking call)
	log.Println("Starting asynq server...")
	if err := s.server.Run(mux); err != nil {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the asynq server
func (s *DocGenService) Shutdown() {
	log.Println("Shutting down asynq server...")
	s.server.Shutdown()
}
