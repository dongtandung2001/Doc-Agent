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
	aiClient *clients.AIClient
	// Add dependencies:
	// - Message queue client (RabbitMQ/Kafka) to enqueue tasks
	// - Cache for analysis results
	// - LLM client for classification
}

func NewAnalysisService(aiClient *clients.AIClient) *AnalysisService {
	return &AnalysisService{
		aiClient: aiClient,
	}
}

// StartCodebaseAnalysis orchestrates the document generation process
func (s *AnalysisService) StartCodebaseAnalysis(
	ctx context.Context,
	req *apiv1.StartCodebaseAnalysisRequest,
) (*apiv1.StartCodebaseAnalysisResponse, error) {
	log.Printf("Starting codebase analysis with project structure: %s", req.ProjectStructure)
	// init chat context
	chatCtx := chatContext.NewContext()
	// Step 1: Classify the repository
	classification, err := pipeline.ClassifyRepo(*chatCtx, s.aiClient)

	// Handle error
	if err != nil {
		log.Printf("Error classifying repository: %v", err)
		return nil, err
	}
	log.Printf("Repository classified as: %s", classification)

	return &apiv1.StartCodebaseAnalysisResponse{Success: true}, nil
}
