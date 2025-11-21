package clients

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	ChatContext "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	promptProcessor "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/utils/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var templateVarRegex = regexp.MustCompile(`\{\{\$([a-zA-Z0-9_]+)\}\}`)

// AIClient wraps the AI service gRPC client
type AIClient struct {
	client apiv1.AIServiceClient
	conn   *grpc.ClientConn
}

// NewAIClient creates a new AI client instance
func NewAIClient(host string, port int) (*AIClient, error) {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service client: %w", err)
	}

	return &AIClient{
		client: apiv1.NewAIServiceClient(conn),
		conn:   conn,
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
	// Prepare the request with the processed prompt
	return c.client.Chat(ctx, req)
}

// HealthCheck checks if the AI service is alive
func (c *AIClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, req)
}

// Close closes the gRPC connection
func (c *AIClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
