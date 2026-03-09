package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	chatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"

	"github.com/hibiken/asynq"
)

// Item represents a documentation item with potential children
type Item struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Prompt      string `json:"prompt"`
	Children    []Item `json:"children,omitempty"`
	Parent      string `json:"parent,omitempty"`
	ProjectType string `json:"projectType,omitempty"`
	CodeFiles   string `json:"code_files,omitempty"`
}

// DocumentStructure represents the root structure
type DocumentStructure struct {
	Items []Item `json:"items"`
}

// stripMarkdownCodeBlock removes markdown code block delimiters from a string
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)

	// Check if string starts with ```
	if strings.HasPrefix(s, "```") {
		// Find the end of the first line (the ```json or just ```)
		firstNewline := strings.Index(s, "\n")
		if firstNewline != -1 {
			s = s[firstNewline+1:]
		}

		// Remove trailing ```
		s = strings.TrimSpace(s)
		if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
	}

	return strings.TrimSpace(s)
}

func parseInstructionJson(instructionJson string) ([]Item, error) {
	// Strip markdown code block delimiters if present
	cleanJson := stripMarkdownCodeBlock(instructionJson)

	var res DocumentStructure
	err := json.Unmarshal([]byte(cleanJson), &res)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}
	return res.Items, nil
}

func flattenItems(items []Item, parentTitle string, projectType string, code_files string) []Item {
	var result []Item
	for _, item := range items {
		item.Parent = parentTitle
		item.CodeFiles = code_files
		item.ProjectType = projectType
		result = append(result, item)
		if len(item.Children) > 0 {
			result = append(result, flattenItems(item.Children, item.Title, projectType, code_files)...)
		}
	}
	return result
}

const (
	TaskTypeDocGenInstruction = "docgen:instruction"
)

func EnqueueInstruction(chatCtx chatContext.ChatContext, instructionJson string, redisClient *clients.RedisClient, dbClient *clients.DatabaseClient) (bool, error) {
	items, err := parseInstructionJson(instructionJson)
	if err != nil {
		return false, err
	}

	code_files, ok := chatCtx.Get("code_files")

	if !ok {
		log.Println("No code files found in chat context")
	}

	log.Printf("DEBUG: code_files for enqueueing: %v", code_files)

	projectType, ok := chatCtx.Get("projectType")
	if !ok {
		log.Println("No project type found in chat context")
	}
	log.Printf("DEBUG: projectType for enqueueing: %v", projectType)

	flatItems := flattenItems(items, "", projectType.(string), code_files.(string))
	log.Printf("Flattened %d items for processing: %+v", len(flatItems), flatItems)

	// Build title→name map so we can resolve parent IDs by parent title
	titleToName := make(map[string]string, len(flatItems))
	for _, item := range flatItems {
		titleToName[item.Title] = item.Name
	}

	const projectID = "1" // TODO: replace with real project ID

	client := redisClient.GetMQClient()
	enqueuedCount := 0

	for i, item := range flatItems {
		// Store section in database before enqueuing
		sectionReq := &apiv1.StoreSectionRequest{
			Id:        item.Title,
			ProjectId: projectID,
			Title:     item.Name,
			Prompt:    item.Prompt,
			Order:     int32(i),
			ParentId:  item.Parent,
		}

		if resp, err := dbClient.StoreSection(context.Background(), sectionReq); err != nil {
			log.Printf("Error storing section %s: %v", item.Title, err)
		} else if !resp.Success {
			log.Printf("Failed to store section %s", item.Title)
		}

		payload, err := json.Marshal(item)
		if err != nil {
			log.Printf("Error marshaling item %s: %v", item.Title, err)
			continue
		}

		task := asynq.NewTask(TaskTypeDocGenInstruction, payload)
		info, err := client.Enqueue(task)
		if err != nil {
			log.Printf("Error enqueuing item %s to task queue: %v", item.Title, err)
			continue
		}

		enqueuedCount++
		log.Printf("Enqueued task: %s (%s) - Queue: %s, ID: %s",
			item.Title, item.Name, info.Queue, info.ID)
	}

	log.Printf("Successfully enqueued %d/%d items to task queue", enqueuedCount, len(flatItems))
	return enqueuedCount == len(flatItems), nil
}
