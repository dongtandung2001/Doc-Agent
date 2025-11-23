package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	ChatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	promptProcessor "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/utils/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var templateVarRegex = regexp.MustCompile(`\{\{\$([a-zA-Z0-9_]+)\}\}`)

// AIClient wraps either gRPC or HTTP chat API client
type AIClient struct {
	// gRPC fields
	client apiv1.AIServiceClient
	conn   *grpc.ClientConn

	// HTTP fields
	httpClient *http.Client
	apiURL     string
	apiKey     string
	model      string

	// Mode flag
	isHTTP bool
}

// NewAIClient creates a new AI client instance
// If host is empty, uses HTTP mode with env vars, otherwise uses gRPC
func NewAIClient(host string, port int) (*AIClient, error) {
	// HTTP mode if no host provided
	if host == "" {
		apiURL := os.Getenv("CHAT_API_URL")
		apiKey := os.Getenv("CHAT_API_KEY")
		model := os.Getenv("CHAT_MODEL")

		if apiURL == "" || apiKey == "" || model == "" {
			return nil, fmt.Errorf("HTTP mode requires CHAT_API_URL, CHAT_API_KEY, CHAT_MODEL env vars")
		}

		fmt.Printf("Successfully created AIClient in HTTP mode with URL: %s\n", apiURL)
		return &AIClient{
			httpClient: &http.Client{},
			apiURL:     apiURL,
			apiKey:     apiKey,
			model:      model,
			isHTTP:     true,
		}, nil
	}

	// gRPC mode
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service client: %w", err)
	}

	fmt.Printf("Successfully created AIClient in gRPC mode with target: %s\n", target)
	return &AIClient{
		client: apiv1.NewAIServiceClient(conn),
		conn:   conn,
		isHTTP: false,
	}, nil
}

// PrepareRequest processes the input before sending to Chat
// Replaces all {{$key}} patterns in the prompt with values from the context
func (c *AIClient) PrepareChatRequest(messages []*apiv1.ChatMessage, chatContext *ChatContext.ChatContext, prompt string, promptProcessingRequire bool) *apiv1.ChatRequest {
	processedPrompt := prompt
	// Replace template variables in the prompt
	if promptProcessingRequire {
		processedPrompt = promptProcessor.ProcessTemplateVariables(prompt, chatContext)
	}

	// Append the processed prompt as a user message
	messages = append(messages, &apiv1.ChatMessage{
		Role:    "user",
		Content: processedPrompt,
	})

	return &apiv1.ChatRequest{
		Messages: messages,
	}
}

// Chat sends a chat request to the AI service
func (c *AIClient) Chat(ctx context.Context, req *apiv1.ChatRequest) (*apiv1.ChatResponse, error) {
	if c.isHTTP {
		return c.chatHTTP(ctx, req)
	}
	return c.client.Chat(ctx, req)
}

// chatHTTP handles HTTP chat requests
func (c *AIClient) chatHTTP(ctx context.Context, req *apiv1.ChatRequest) (*apiv1.ChatResponse, error) {
	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "fs_read",
				"description": "Read file contents with optional line range",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path to read",
						},
						"start_line": map[string]interface{}{
							"type":        "integer",
							"description": "Starting line (1-indexed)",
						},
						"end_line": map[string]interface{}{
							"type":        "integer",
							"description": "Ending line (1-indexed, inclusive)",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		// Add more tools here if needed
	}
	// Convert proto messages to JSON format
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := make([]msg, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = msg{Role: m.Role, Content: m.Content}
	}

	reqBody := struct {
		Model      string                   `json:"model"`
		Messages   []msg                    `json:"messages"`
		Tools      []map[string]interface{} `json:"tools"`
		ToolChoice string                   `json:"tool_choice,omitempty"`
	}{
		Model:      c.model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	jsonData, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	return &apiv1.ChatResponse{
		Content: result.Choices[0].Message.Content,
	}, nil
}

// HealthCheck checks if the AI service is alive
func (c *AIClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	if c.isHTTP {
		// Simple health check for HTTP - try a minimal chat request
		testReq := &apiv1.ChatRequest{
			Messages: []*apiv1.ChatMessage{{Role: "user", Content: "test"}},
		}
		_, err := c.chatHTTP(ctx, testReq)
		if err != nil {
			return &apiv1.HealthCheckResponse{IsAlive: false}, err
		}
		return &apiv1.HealthCheckResponse{IsAlive: true}, nil
	}
	return c.client.HealthCheck(ctx, req)
}

// Close closes the connection
func (c *AIClient) Close() error {
	if c.isHTTP {
		c.httpClient.CloseIdleConnections()
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

