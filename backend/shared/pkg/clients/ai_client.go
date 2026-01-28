package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	ChatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	promptProcessor "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/utils/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Context keys for metadata
type ContextKey string

const (
	AgenticMode    ContextKey = "agentic_chat"
	HTTPTimeoutKey ContextKey = "http_timeout"
	ToolChoiceKey  ContextKey = "tool_choice"
	ToolRequire    ContextKey = "tool_require"
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
	Content      string
	ToolCalls    []*apiv1.ToolCall
	FinishReason string
}

// ToolUse represents a parsed tool call ready for execution
// Matches state.rs ToolUse structure
type ToolUse struct {
	ID   string                 // Tool call ID
	Name string                 // Tool name
	Args map[string]interface{} // Parsed arguments
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
			Timeout: 10 * time.Minute,
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
							"description": "File path to read.",
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
// Returns "required" to force tool usage
func (c *AIClient) GetToolChoice() string {
	return "auto"
}

// ExecuteTool executes a tool with the given name and file requests
// For fs_read, it calls gatewayClient.RequestFileContent with the batch of files
func (c *AIClient) ExecuteTool(toolName string, fileRequests []*apiv1.FileReadArg, gatewayClient *GatewayClient) ([]*apiv1.FileReadResult, error) {
	if toolName == "fs_read" {
		fmt.Printf("[DEBUG] Executing fs_read tool with %d files\n", len(fileRequests))

		// Call gateway to request file contents in batch
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		req := &apiv1.RequestFileContentRequest{
			Args: fileRequests,
		}
		response, err := gatewayClient.RequestFileContent(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to request file content: %w", err)
		}

		// Return the FileReadResult array directly
		log.Print("DEBUG: FileReadResults: ", response.Results)
		return response.Results, nil
	}

	return nil, fmt.Errorf("unknown tool: %s", toolName)
}

// Chat sends a chat request to the AI service (main orchestrator)
// If context contains AgenticMode=true, it will run an agentic loop with tool execution
func (c *AIClient) Chat(ctx context.Context, req *apiv1.ChatRequest, gatewayClient *GatewayClient) (*apiv1.ChatResponse, error) {
	// Check if agentic mode is enabled
	isAgentic, ok := ctx.Value(AgenticMode).(bool)
	log.Printf("AgenticMode: %v", isAgentic)

	if !ok || !isAgentic {
		// Non-agentic mode: single request-response
		log.Printf("Non-agentic mode: sending single chat request")
		parsed, err := c.sendMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		return &apiv1.ChatResponse{Content: parsed.Content}, nil
	}

	log.Printf("Agentic mode: starting agentic loop")

	// Agentic mode: loop with tool execution
	// Work directly with req.Messages to allow caller to see conversation history

	maxIterations := 15 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++
		log.Printf("[DEBUG] Agentic loop iteration %d\n", iteration)

		// Send message using protocol-specific helper
		chatReq := &apiv1.ChatRequest{Messages: req.Messages}
		parsed, err := c.sendMessage(ctx, chatReq)
		if err != nil {
			return nil, fmt.Errorf("API error in iteration %d: %w", iteration, err)
		}

		// get finish reason
		log.Printf("[DEBUG] Finish reason: %s\n", parsed.FinishReason)
		// If no tool calls, we're done
		if len(parsed.ToolCalls) == 0 {
			log.Printf("[DEBUG] No tool calls, returning response\n")
			return &apiv1.ChatResponse{Content: parsed.Content}, nil
		}

		req.Messages = append(req.Messages, &apiv1.ChatMessage{
			Role:      "assistant",
			Content:   parsed.Content,
			ToolCalls: parsed.ToolCalls,
		})

		fmt.Printf("[DEBUG] Executing %d tools\n", len(parsed.ToolCalls))
		fmt.Printf("[DEBUG] ToolCalls: %+v\n", parsed.ToolCalls)

		// Step 1: Parse tool calls and build FileReadArg array directly
		fileRequests := make([]*apiv1.FileReadArg, 0, len(parsed.ToolCalls))

		for _, toolCall := range parsed.ToolCalls {
			// Parse arguments JSON string to map
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			// Extract path (required)
			path, ok := args["path"].(string)
			if !ok {
				fmt.Printf("[DEBUG] Warning: missing path for tool ID %s\n", toolCall.Id)
				continue
			}

			// Build FileReadArg using proto type
			fileArg := &apiv1.FileReadArg{
				Id:   toolCall.Id, // proto field 1
				Path: path,        // proto field 2
			}

			fileRequests = append(fileRequests, fileArg)
		}

		// Step 2: Execute all tools in batch via ExecuteTool
		fmt.Printf("[DEBUG] Executing %d fs_read tools in batch\n", len(fileRequests))
		fileResults, err := c.ExecuteTool("fs_read", fileRequests, gatewayClient)

		// Step 3: Add file results to messages
		if err != nil {
			return nil, fmt.Errorf("tool execution error: %w", err)
		} else {
			// Iterate through FileReadResult array and add to messages
			log.Print("DEBUG: FileReadResults: ", fileResults)
			for _, result := range fileResults {
				req.Messages = append(req.Messages, &apiv1.ChatMessage{
					Role:       "tool",
					Content:    result.Content,
					ToolCallId: &result.Id,
				})
			}
		}

		log.Print("DEBUG: CurrentMessages: ", req.Messages)

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
	type toolCallFunction struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type toolCall struct {
		Id       string           `json:"id"`
		Type     string           `json:"type"`
		Function toolCallFunction `json:"function"`
	}
	type msg struct {
		Role       string      `json:"role"`
		Content    string      `json:"content"`
		ToolCallId string      `json:"tool_call_id,omitempty"`
		ToolCalls  []toolCall  `json:"tool_calls,omitempty"`
	}
	messages := make([]msg, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = msg{Role: m.Role, Content: m.Content}
		// Add tool_call_id for tool response messages
		if m.ToolCallId != nil && *m.ToolCallId != "" {
			messages[i].ToolCallId = *m.ToolCallId
		}
		// Add tool_calls for assistant messages
		if len(m.ToolCalls) > 0 {
			messages[i].ToolCalls = make([]toolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				messages[i].ToolCalls[j] = toolCall{
					Id:   tc.Id,
					Type: tc.Type,
					Function: toolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}

	tools := []map[string]interface{}{}
	toolChoice := "none"
	// Get tools from client
	isToolRequire, ok := ctx.Value(ToolRequire).(bool)

	if !ok {
		isToolRequire = false
	}

	if isToolRequire {
		tools = c.GetTools()
		toolChoice = c.GetToolChoice()
	}

	log.Printf("DEBUG: isToolRequire: %v, tools: %+v, toolChoice: %s", isToolRequire, tools, toolChoice)

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
	log.Printf("DEBUG: reqBody: %+v", httpReq)

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
				Content   string            `json:"content"`
				ToolCalls []*apiv1.ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	log.Printf("DEBUG: Decoding API response: %s", resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	log.Printf("DEBUG: API response: %+v", result)

	return &ParsedResponse{
		Content:      result.Choices[0].Message.Content,
		ToolCalls:    result.Choices[0].Message.ToolCalls,
		FinishReason: result.Choices[0].FinishReason,
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
