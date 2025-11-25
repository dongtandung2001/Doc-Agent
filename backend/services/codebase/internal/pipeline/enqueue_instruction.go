package pipeline

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	"github.com/hibiken/asynq"
)

// Item represents a documentation item with potential children
type Item struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Prompt   string `json:"prompt"`
	Children []Item `json:"children,omitempty"`
	Parent   string `json:"parent,omitempty"`
}

// DocumentStructure represents the root structure
type DocumentStructure struct {
	Items []Item `json:"items"`
}

func parseInstructionJson(instructionJson string) ([]Item, error) {
	var res DocumentStructure
	err := json.Unmarshal([]byte(instructionJson), &res)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}
	return res.Items, nil
}

func flattenItems(items []Item, parentTitle string) []Item {
	var result []Item
	for _, item := range items {
		item.Parent = parentTitle
		result = append(result, item)
		if len(item.Children) > 0 {
			result = append(result, flattenItems(item.Children, item.Title)...)
		}
	}
	return result
}

const (
	TaskTypeDocGenInstruction = "docgen:instruction"
)

func EnqueueInstruction(instructionJson string, redisClient *clients.RedisClient) (bool, error) {
	items, err := parseInstructionJson(instructionJson)
	if err != nil {
		return false, err
	}

	flatItems := flattenItems(items, "")
	log.Printf("Flattened %d items for processing: %+v", len(flatItems), flatItems)

	client := redisClient.GetMQClient()
	enqueuedCount := 0

	for _, item := range flatItems {
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
