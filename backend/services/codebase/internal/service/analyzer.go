package service

import (
	"context"
	"log"

	pipeline "github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/pipeline"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

type AnalysisService struct {
	aiClient      *clients.AIClient
	gatewayClient *clients.GatewayClient
	redisClient   *clients.RedisClient
	// Add dependencies:
	// - Message queue client (RabbitMQ/Kafka) to enqueue tasks
	// - Cache for analysis results
	// - LLM client for classification
}

func NewAnalysisService(aiClient *clients.AIClient, gatewayClient *clients.GatewayClient, redisClient *clients.RedisClient) *AnalysisService {
	return &AnalysisService{
		aiClient:      aiClient,
		gatewayClient: gatewayClient,
		redisClient:   redisClient,
	}
}

// StartCodebaseAnalysis orchestrates the document generation process
func (s *AnalysisService) StartCodebaseAnalysis(
	ctx context.Context,
	req *apiv1.StartCodebaseAnalysisRequest,
) (*apiv1.StartCodebaseAnalysisResponse, error) {
	log.Printf("Starting codebase analysis with project structure: %s", req.ProjectStructure)
	log.Printf("Readme content: %s", req.ReadmeContent)
	// init chat context
	chatCtx := chatContext.NewContext()
	// get req
	projectStructure := req.GetProjectStructure()
	readmeContent := req.GetReadmeContent()
	// set chat context
	chatCtx.Set("category", projectStructure)
	chatCtx.Set("readme", readmeContent)
	// Step 1: Classify the repository
	classification, err := pipeline.ClassifyRepo(*chatCtx, s.aiClient, s.gatewayClient)

	// Handle error
	if err != nil {
		log.Printf("Error classifying repository: %v", err)
		return nil, err
	}

	// Clear the chatContext for next use
	chatCtx.Clear()

	// set the classification result to chat context
	chatCtx.Set("classification", classification)
	chatCtx.Set("code_files", projectStructure)
	// Step 2: Generate instructions based on classification
	instructions, err2 := pipeline.GenerateInstruction(*chatCtx, s.aiClient, s.gatewayClient)

	// Handle error
	if err2 != nil {
		log.Printf("Error generating instructions: %v", err2)
		return nil, err2
	}
	// Step 3: Enqueue instruction processing task

	success, err3 := pipeline.EnqueueInstruction(*chatCtx, instructions, s.redisClient)
	if err3 != nil {
		log.Printf("Error enqueueing instruction processing: %v", err3)
		return nil, err3
	}

	if !success {
		log.Printf("Not all instructions were enqueued successfully")
		return &apiv1.StartCodebaseAnalysisResponse{Success: false}, nil
	}

	return &apiv1.StartCodebaseAnalysisResponse{Success: true}, nil
}
