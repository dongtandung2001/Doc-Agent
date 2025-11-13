package service

import (
	"context"
	"encoding/json"
	"log"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
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

	// 1. Parse the project structure JSON
	var projectStructure map[string]interface{}
	if err := json.Unmarshal([]byte(req.ProjectStructure), &projectStructure); err != nil {
		return &apiv1.StartCodebaseAnalysisResponse{Success: false}, err
	}

	// 2. Analyze codebase structure and classify the project
	//    - Send project structure + overview to LLM
	//    - LLM generates documentation sections, subsections, and instructions
	//    Example instruction:
	//    {
	//      "Id": "getting-started",
	//      "Title": "Getting Started",
	//      "Instruction": "Help users understand and start using the project",
	//      "IsCompleted": false,
	//      "Order": 0
	//    }

	// 3. Enqueue all instructions to Message Queue (RabbitMQ/Kafka)
	//    for Document Generation service to process in parallel
	//
	//    TODO: Implement message queue producer
	//    for _, instruction := range instructions {
	//        messageQueue.Publish("doc-generation-tasks", instruction)
	//    }

	log.Println("Codebase analysis started and tasks enqueued successfully")

	return &apiv1.StartCodebaseAnalysisResponse{Success: true}, nil
}
