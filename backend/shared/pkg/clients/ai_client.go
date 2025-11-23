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
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	ChatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	promptProcessor "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/utils/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var templateVarRegex = regexp.MustCompile(`\{\{\$([a-zA-Z0-9_]+)\}\}`)

// Context keys for metadata
type ContextKey string

const (
	AgenticMode    ContextKey = "agentic_chat"
	HTTPTimeoutKey ContextKey = "http_timeout"
)

// ToolCall represents a tool call from the API response
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function details in a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ParsedResponse represents parsed API response with tools support
type ParsedResponse struct {
	Content   string
	ToolCalls []ToolCall
}

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

		// Use a longer timeout for AI API calls (5 minutes default)
		httpClient := &http.Client{
			Timeout: 5 * time.Minute,
		}

		fmt.Printf("Successfully created AIClient in HTTP mode with URL: %s\n", apiURL)
		return &AIClient{
			httpClient: httpClient,
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

// GetTools returns the list of available tools for the AI
// Override this method to provide custom tools
func (c *AIClient) GetTools() []map[string]interface{} {
	return []map[string]interface{}{
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
	}
}

// GetToolChoice returns the tool choice strategy
// Override this method to customize tool selection behavior
// Returns "auto" by default
func (c *AIClient) GetToolChoice() string {
	return "auto"
}

// ExecuteTool executes a tool with the given name and arguments
// This method should be overridden by embedding AIClient in a custom struct
// Default implementation returns an error
func (c *AIClient) ExecuteTool(toolName string, arguments map[string]interface{}, gatewayClient *GatewayClient) (string, error) {
	if toolName == "fs_read" {
		// Extract arguments
		path, _ := arguments["path"].(string)
		// Tool call id
		tool_call_id, _ := arguments["tool_call_id"].(string)
		fmt.Printf("[DEBUG] Executing fs_read tool (ID: %s) for path: %s\n", tool_call_id, path)

		

		return "", nil
	}
	return "", fmt.Errorf("tool execution not implemented: override ExecuteTool method")
}

// Chat sends a chat request to the AI service (main orchestrator)
// If context contains AgenticMode=true, it will run an agentic loop with tool execution
func (c *AIClient) Chat(ctx context.Context, req *apiv1.ChatRequest, gatewayClient *GatewayClient) (*apiv1.ChatResponse, error) {
	// Check if agentic mode is enabled
	isAgentic, _ := ctx.Value(AgenticMode).(bool)

	if !isAgentic {
		// Non-agentic mode: single request-response
		parsed, err := c.sendMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		return &apiv1.ChatResponse{Content: parsed.Content}, nil
	}

	// Agentic mode: loop with tool execution

	// Clone messages to avoid modifying the original slice
	currentMessages := make([]*apiv1.ChatMessage, len(req.Messages))
	copy(currentMessages, req.Messages)

	maxIterations := 10 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++
		fmt.Printf("[DEBUG] Agentic loop iteration %d\n", iteration)

		// Send message using protocol-specific helper
		chatReq := &apiv1.ChatRequest{Messages: currentMessages}
		parsed, err := c.sendMessage(ctx, chatReq)
		if err != nil {
			return nil, fmt.Errorf("API error in iteration %d: %w", iteration, err)
		}

		// If no tool calls, we're done
		if len(parsed.ToolCalls) == 0 {
			fmt.Printf("[DEBUG] No tool calls, returning response\n")
			return &apiv1.ChatResponse{Content: parsed.Content}, nil
		}

		// Add assistant message with tool calls to conversation
		if parsed.Content != "" {
			currentMessages = append(currentMessages, &apiv1.ChatMessage{
				Role:    "assistant",
				Content: parsed.Content,
			})
		}

		fmt.Printf("[DEBUG] Executing %d tools\n", len(parsed.ToolCalls))

		// Execute each tool and collect results
		for _, toolCall := range parsed.ToolCalls {
			fmt.Printf("[DEBUG] Executing tool: %s (ID: %s)\n", toolCall.Function.Name, toolCall.ID)

			// Parse arguments JSON string to map
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			// Execute the tool using the client's ExecuteTool method
			result, err := c.ExecuteTool(toolCall.Function.Name, args, gatewayClient)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %s", err.Error())
			}

			// Add tool result as a message
			toolResultMsg := &apiv1.ChatMessage{
				Role:    "tool",
				Content: fmt.Sprintf("[Tool: %s, ID: %s]\n%s", toolCall.Function.Name, toolCall.ID, result),
			}
			currentMessages = append(currentMessages, toolResultMsg)
		}

		// Continue loop with updated messages
	}

	return nil, fmt.Errorf("exceeded maximum iterations (%d) in agentic loop", maxIterations)
}

// sendMessage routes to the appropriate protocol-specific helper (HTTP or gRPC)
func (c *AIClient) sendMessage(ctx context.Context, req *apiv1.ChatRequest) (*ParsedResponse, error) {
	if c.isHTTP {
		return c.sendHTTP(ctx, req)
	}
	return c.sendGRPC(ctx, req)
}

// sendHTTP handles HTTP-specific message sending
func (c *AIClient) sendHTTP(ctx context.Context, req *apiv1.ChatRequest) (*ParsedResponse, error) {
	// Convert proto messages to JSON format
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := make([]msg, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = msg{Role: m.Role, Content: m.Content}
	}

	// Get tools from client
	tools := c.GetTools()
	toolChoice := c.GetToolChoice()

	// Build request body
	reqBody := struct {
		Model      string                   `json:"model"`
		Messages   []msg                    `json:"messages"`
		Tools      []map[string]interface{} `json:"tools,omitempty"`
		ToolChoice string                   `json:"tool_choice,omitempty"`
	}{
		Model:      c.model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: toolChoice,
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
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	return &ParsedResponse{
		Content:   result.Choices[0].Message.Content,
		ToolCalls: result.Choices[0].Message.ToolCalls,
	}, nil
}

// sendGRPC handles gRPC-specific message sending
func (c *AIClient) sendGRPC(ctx context.Context, req *apiv1.ChatRequest) (*ParsedResponse, error) {
	// Call gRPC endpoint
	resp, err := c.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert to ParsedResponse
	// TODO: Parse tool_calls from gRPC response when proto supports it
	return &ParsedResponse{
		Content:   resp.Content,
		ToolCalls: nil, // Will be populated when proto is updated
	}, nil
}

// HealthCheck checks if the AI service is alive
func (c *AIClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	if c.isHTTP {
		// Simple health check for HTTP - try a minimal chat request
		testReq := &apiv1.ChatRequest{
			Messages: []*apiv1.ChatMessage{{Role: "user", Content: "test"}},
		}
		_, err := c.sendHTTP(ctx, testReq)
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
