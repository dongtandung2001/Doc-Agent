package pipeline

import (
	"log"
	"os"
)

func classify_repo(projectStructure string) (string, error) {
	// Construct the prompt for classification
	prompt, err := os.ReadFile("../prompts/classfication_prompt.md")

	if err != nil {
		log.Printf("Error reading prompt file: %v", err)
		return "", err
	}

	log.Printf("prompt: %s", prompt)

	return "", nil
}
