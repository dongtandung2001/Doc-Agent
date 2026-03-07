package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/pipeline"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	"github.com/hibiken/asynq"
)

const (
	TaskTypeDocGenInstruction = "docgen:instruction"
)

type Item struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Prompt      string `json:"prompt"`
	Children    []Item `json:"children,omitempty"`
	Parent      string `json:"parent,omitempty"`
	ProjectType string `json:"projectType,omitempty"`
	CodeFiles   string `json:"code_files,omitempty"`
}

// TaskHandler handles all task processing for the docgen service
type TaskHandler struct {
	aiClient      *clients.AIClient
	gatewayClient *clients.GatewayClient
	dbClient      *clients.DatabaseClient
}

// NewTaskHandler creates a new task handler with necessary dependencies
func NewTaskHandler(aiClient *clients.AIClient, gatewayClient *clients.GatewayClient, dbClient *clients.DatabaseClient) *TaskHandler {
	return &TaskHandler{
		aiClient:      aiClient,
		gatewayClient: gatewayClient,
		dbClient:      dbClient,
	}
}

// HandleDocGenInstruction processes a documentation generation instruction task
func (h *TaskHandler) HandleDocGenInstruction(ctx context.Context, task *asynq.Task) error {
	var item Item
	if err := json.Unmarshal(task.Payload(), &item); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	log.Printf("Processing docgen task: Title=%s, Name=%s, Parent=%s", item.Title, item.Name, item.Parent)
	log.Printf("Payload: %+v", item)
	// Init ChatContext
	chatCtx := chatContext.NewContext()
	// Set item-specific context
	chatCtx.Set("title", item.Title)
	chatCtx.Set("name", item.Name)
	chatCtx.Set("prompt", item.Prompt)
	chatCtx.Set("projectType", item.ProjectType)
	chatCtx.Set("code_files", item.CodeFiles)
	// Call AI and Gateway clients to generate documentation
	_, err := pipeline.GenerateDocumentation(*chatCtx, h.aiClient, h.gatewayClient, h.dbClient)
	if err != nil {
		log.Printf("Error generating documentation for %s: %v", item.Title, err)
		return err
	}

	log.Printf("Successfully processed docgen task for: %s", item.Title)
	return nil
}

// RegisterHandlers returns a ServeMux with all task handlers registered
func (h *TaskHandler) RegisterHandlers() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeDocGenInstruction, h.HandleDocGenInstruction)
	return mux
}
