package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

func GenerateDocumentation(chatContext chatContext.ChatContext, aiClient *clients.AIClient, gatewayClient *clients.GatewayClient) (string, error) {

	title, ok := chatContext.Get("title")
	if !ok {
		log.Println("No title found in chat context")
		return "", nil
	}

	// prompt, ok := chatContext.Get("prompt")
	// if !ok {
	// 	log.Println("No prompt found in chat context")
	// 	return "", nil
	// }

	// code_files, ok := chatContext.Get("code_files")
	// if !ok {
	// 	log.Println("No code files found in chat context")
	// 	return "", nil
	// }

	// projectType, ok := chatContext.Get("projectType")
	// if !ok {
	// 	log.Println("No project type found in chat context")
	// 	return "", nil
	// }

	base_prompt, err := os.ReadFile("internal/prompts/generate_doc.md")
	if err != nil {
		log.Println("Failed to read base prompt:", err)
		return "", err
	}

	// init messages array
	// log.Printf("Generating documentation for project type: %s with code_files=%v, title=%s, prompt=%s", projectType, code_files, title, prompt)
	messages := []*apiv1.ChatMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	req := aiClient.PrepareChatRequest(messages, &chatContext, string(base_prompt), true)
	// set agentic mode
	ctx = context.WithValue(ctx, clients.AgenticMode, true)
	ctx = context.WithValue(ctx, clients.ToolRequire, true)
	messages = req.Messages // Update messages to include the new message added by PrepareChatRequest
	generated_doc, err := aiClient.Chat(ctx, req, gatewayClient)
	if err != nil {
		log.Println("Error during generating documentation:", err)
		return "", err
	}

	// store generated documentation local storage
	docsDir := "/generated_docs"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Println("Failed to create generated_docs directory:", err)
		return "", err
	}

	// Sanitize title for filename
	sanitizedTitle := sanitizeFilename(title.(string))
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.md", sanitizedTitle, timestamp)
	filePath := filepath.Join(docsDir, filename)

	// Write documentation to file
	if err := os.WriteFile(filePath, []byte(generated_doc.Content), 0644); err != nil {
		log.Println("Failed to write documentation to file:", err)
		return "", err
	}

	log.Printf("Documentation successfully saved to: %s", filePath)
	return fmt.Sprintf("Success: Documentation saved to %s", filePath), nil
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	// Remove invalid characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	name = reg.ReplaceAllString(name, "")
	// Limit length to 50 characters
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}
