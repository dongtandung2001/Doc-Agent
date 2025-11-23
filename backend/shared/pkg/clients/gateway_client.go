package clients

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1/protov1connect"
)

// GatewayClient wraps the Gateway service Connect-Go client
type GatewayClient struct {
	client     protov1connect.GatewayServiceClient
	httpClient *http.Client
}

// NewGatewayClient creates a new Gateway client instance using Connect-Go
func NewGatewayClient(host string, port int) (*GatewayClient, error) {
	// Construct base URL for Connect-Go (uses HTTP, not gRPC)
	baseURL := fmt.Sprintf("http://%s", net.JoinHostPort(host, strconv.Itoa(port)))

	// Create HTTP client for Connect-Go
	httpClient := &http.Client{}

	// Create Connect-Go client
	client := protov1connect.NewGatewayServiceClient(httpClient, baseURL)

	return &GatewayClient{
		client:     client,
		httpClient: httpClient,
	}, nil
}

// GetClient returns the underlying Connect-Go client for direct access
func (c *GatewayClient) GetClient() protov1connect.GatewayServiceClient {
	return c.client
}

func (c *GatewayClient) RequestFileContent(ctx context.Context, req *apiv1.RequestFileContentRequest) (*apiv1.RequestFileContentResponse, error) {
	// Wrap request in connect.Request
	connectReq := connect.NewRequest(req)

	// Call Connect-Go client
	resp, err := c.client.RequestFileContent(ctx, connectReq)
	if err != nil {
		return nil, err
	}

	// Return unwrapped response message
	return resp.Msg, nil
}

// HealthCheck checks if the Gateway service is alive
func (c *GatewayClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	// Wrap request in connect.Request
	connectReq := connect.NewRequest(req)

	// Call Connect-Go client
	resp, err := c.client.HealthCheck(ctx, connectReq)
	if err != nil {
		return nil, err
	}

	// Return unwrapped response message
	return resp.Msg, nil
}

// Close closes the HTTP client
func (c *GatewayClient) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}
