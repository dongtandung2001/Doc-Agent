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

func ClassifyRepo(chatContext ctx.ChatContext, aiClient *clients.AIClient) (string, error) {
	// Construct the prompt for classification
	prompt, err := os.ReadFile("../prompts/classfication_prompt.md")

	if err != nil {
		log.Printf("Error reading prompt file: %v", err)
		return "", err
	}

	messages := []*apiv1.ChatMessage{}

	chatRequest := aiClient.PrepareChatRequest(messages, &chatContext, string(prompt), true)

	log.Printf("prompt: %s", prompt)

	// In your gRPC client call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	classfication, err := aiClient.Chat(ctx, chatRequest)

	log.Printf("RepoClassification: classification: %s", classfication.Content)

	return classfication.Content, err
}
