package pipeline

import (
	"context"
	"log"
	"os"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	ctx "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

func ClassifyRepo(chatContext ctx.ChatContext, aiClient *clients.AIClient, gatewayClient *clients.GatewayClient) (string, error) {
	// Construct the prompt for classification
	prompt, err := os.ReadFile("internal/prompts//generate_classfication.md")

	if err != nil {
		log.Printf("Error reading prompt file: %v", err)
		return "", err
	}

	messages := []*apiv1.ChatMessage{}

	chatRequest := aiClient.PrepareChatRequest(messages, &chatContext, string(prompt), true)

	log.Printf("Final repoClassification prompt: %s", chatRequest.Messages[0])

	// In your gRPC client call
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, clients.AgenticMode, false)
	ctx = context.WithValue(ctx, clients.ToolRequire, false)
	classfication, err := aiClient.Chat(ctx, chatRequest, gatewayClient)

	if err != nil {
		log.Printf("Error calling AI Chat: %v", err)
		return "", err
	}

	log.Printf("RepoClassification: classification: %s", classfication)

	return classfication.Content, nil
}
