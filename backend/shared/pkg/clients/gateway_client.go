package clients

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"google.golang.org/grpc"
)

// GatewayClient wraps the Gateway service gRPC client
type GatewayClient struct {
	client apiv1.GatewayServiceClient
	conn   *grpc.ClientConn
}

// NewGatewayClient creates a new Gateway client instance
func NewGatewayClient(host string, port int) (*GatewayClient, error) {
	conn, err := newGRPCConnection(host, port)
	if err != nil {
		return nil, err
	}

	return &GatewayClient{
		client: apiv1.NewGatewayServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetClient returns the underlying gRPC client for direct access
func (c *GatewayClient) GetClient() apiv1.GatewayServiceClient {
	return c.client
}

func (c *GatewayClient) RequestFileContent(ctx context.Context, req *apiv1.RequestFileContentRequest) (*apiv1.RequestFileContentResponse, error) {
	return c.client.RequestFileContent(ctx, req)
}

// HealthCheck checks if the Gateway service is alive
func (c *GatewayClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, req)
}

// Close closes the gRPC connection
func (c *GatewayClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
