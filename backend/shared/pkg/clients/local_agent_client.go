package clients

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"google.golang.org/grpc"
)

// LocalAgentClient wraps the Local Agent service gRPC client (Rust service)
type LocalAgentClient struct {
	client apiv1.LocalAgentServiceClient
	conn   *grpc.ClientConn
}

// NewLocalAgentClient creates a new Local Agent client instance
func NewLocalAgentClient(host string, port int) (*LocalAgentClient, error) {
	conn, err := newGRPCConnection(host, port)
	if err != nil {
		return nil, err
	}

	return &LocalAgentClient{
		client: apiv1.NewLocalAgentServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetClient returns the underlying gRPC client for direct access
func (c *LocalAgentClient) GetClient() apiv1.LocalAgentServiceClient {
	return c.client
}

func (c *LocalAgentClient) RequestFileContent(ctx context.Context, req *apiv1.RequestFileContentRequest) (*apiv1.RequestFileContentResponse, error) {
	return c.client.RequestFileContent(ctx, req)
}

// HealthCheck checks if the Local Agent service is alive
func (c *LocalAgentClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, req)
}

// Close closes the gRPC connection
func (c *LocalAgentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
