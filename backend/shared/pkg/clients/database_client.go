package clients

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"google.golang.org/grpc"
)

// DatabaseClient wraps the Database service gRPC client
type DatabaseClient struct {
	client apiv1.DatabaseServiceClient
	conn   *grpc.ClientConn
}

// NewDatabaseClient creates a new Database client instance
func NewDatabaseClient(host string, port int) (*DatabaseClient, error) {
	conn, err := newGRPCConnection(host, port)
	if err != nil {
		return nil, err
	}

	return &DatabaseClient{
		client: apiv1.NewDatabaseServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetClient returns the underlying gRPC client for direct access
func (c *DatabaseClient) GetClient() apiv1.DatabaseServiceClient {
	return c.client
}

// HealthCheck checks if the Database service is alive
func (c *DatabaseClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, req)
}

// Close closes the gRPC connection
func (c *DatabaseClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
