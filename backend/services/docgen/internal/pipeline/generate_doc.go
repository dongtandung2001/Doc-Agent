package pipeline

import (
	"log"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

func GenerateDocumentation(chatContext chatContext.ChatContext, aiClient *clients.AIClient, gatewayClient *clients.GatewayClient) (string, error) {

	// base_prompt, err := os.ReadFile("internal/prompts/generate_doc.md")
	// if err != nil {
	// 	log.Println("Failed to read base prompt:", err)
	// 	return "", err
	// }

	// // init messages array
	// messages := []*apiv1.ChatMessage{}

	// req := aiClient.PrepareChatRequest(messages, &chatContext, string(base_prompt), true)

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

	log.Printf("Generating documentation for project type: %s with code_files=%v, title=%s, prompt=%s", projectType, code_files, title, prompt)

	// log.Printf("Final doc generation prompt: %s", req.Messages[0])

	return "Success", nil
}
