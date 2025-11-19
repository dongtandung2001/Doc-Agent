package pipeline

import (
	"log"
	"os"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	ctx "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

func GenerateInstruction(chatContext ctx.ChatContext, aiClient *clients.AIClient) (string, error) {
	// Construct the prompt for classification
	prompt, err := os.ReadFile("../prompts/generate_instructions.md")

	if err != nil {
		log.Printf("Error reading prompt file: %v", err)
		return "", err
	}

	log.Printf("prompt: %s", prompt)

	return "", nil
}
