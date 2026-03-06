package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

func GenerateDocumentation(chatContext chatContext.ChatContext, aiClient *clients.AIClient, gatewayClient *clients.GatewayClient, dbClient *clients.DatabaseClient) (string, error) {

	title, ok := chatContext.Get("title")
	if !ok {
		log.Println("No title found in chat context")
		return "", nil
	}

	prompt, ok := chatContext.Get("prompt")
	if !ok {
		log.Println("No prompt found in chat context")
		return "", nil
	}

	name, ok := chatContext.Get("name")
	if !ok {
		log.Println("No name found in chat context")
		return "", nil
	}

	code_files, ok := chatContext.Get("code_files")
	if !ok {
		log.Println("No code files found in chat context")
		return "", nil
	}

	projectType, ok := chatContext.Get("projectType")
	if !ok {
		log.Println("No project type found in chat context")
		return "", nil
	}

	base_prompt, err := os.ReadFile("internal/prompts/generate_doc.md")
	if err != nil {
		log.Println("Failed to read base prompt:", err)
		return "", err
	}

	log.Printf("Generating documentation for project type: %s with code_files=%v, title=%s, prompt=%s", projectType, code_files, title, prompt)
	messages := []*apiv1.ChatMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	req := aiClient.PrepareChatRequest(messages, &chatContext, string(base_prompt), true)
	ctx = context.WithValue(ctx, clients.AgenticMode, true)
	ctx = context.WithValue(ctx, clients.ToolRequire, true)
	messages = req.Messages
	generated_doc, err := aiClient.Chat(ctx, req, gatewayClient)
	if err != nil {
		log.Println("Error during generating documentation:", err)
		return "", err
	}

	// Store generated documentation in the database
	timestamp := time.Now().Format("20060102_150405")
	fileID := fmt.Sprintf("%s_%s", name.(string), timestamp)

	resp, err := dbClient.StoreDocument(context.Background(), &apiv1.StoreDocumentRequest{
		Id:          fileID,
		ProjectId:   "1", // TODO: replace with real project ID
		DocumentId:  name.(string),
		Title:       title.(string),
		Description: prompt.(string),
		Content:     generated_doc.Content,
	})
	if err != nil {
		log.Printf("Error storing document %s: %v", title, err)
		return "", err
	}
	if !resp.Success {
		log.Printf("Failed to store document %s", title)
		return "", fmt.Errorf("failed to store document: %s", title)
	}

	log.Printf("Documentation successfully stored for: %s (id=%s)", title, fileID)
	return fmt.Sprintf("Success: Documentation stored with id=%s", fileID), nil
}
