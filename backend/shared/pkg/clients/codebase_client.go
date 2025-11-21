package clients

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"google.golang.org/grpc"
)

// CodebaseClient wraps the Codebase Analysis service gRPC client
type CodebaseClient struct {
	client apiv1.CodebaseAnalysisServiceClient
	conn   *grpc.ClientConn
}

// NewCodebaseClient creates a new Codebase Analysis client instance
func NewCodebaseClient(host string, port int) (*CodebaseClient, error) {
	conn, err := newGRPCConnection(host, port)
	if err != nil {
		return nil, err
	}

	return &CodebaseClient{
		client: apiv1.NewCodebaseAnalysisServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetClient returns the underlying gRPC client for direct access
func (c *CodebaseClient) GetClient() apiv1.CodebaseAnalysisServiceClient {
	return c.client
}

// StartCodebaseAnalysis forwards the request to the Codebase Analysis service
func (c *CodebaseClient) StartCodebaseAnalysis(ctx context.Context, req *apiv1.StartCodebaseAnalysisRequest) (*apiv1.StartCodebaseAnalysisResponse, error) {
	return c.client.StartCodebaseAnalysis(ctx, req)
}

// HealthCheck checks if the Codebase Analysis service is alive
func (c *CodebaseClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, req)
}

// Close closes the gRPC connection
func (c *CodebaseClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
