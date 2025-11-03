package clients

import (
	"fmt"
	"net"
	"strconv"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/gen/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServiceClient[C any](host string, port int, newClientService func(grpc.ClientConnInterface) C) (C, *grpc.ClientConn, error) {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(
		target,
		opts...,
	)
	if err != nil {
		var zero C
		return zero, nil, fmt.Errorf("failed to create service client: %w", err)
	}

	return newClientService(conn), conn, nil
}

// NewLocalAgentClient creates a gRPC client for the Local Agent service (Rust)
func NewLocalAgentClient(host string, port int) (apiv1.LocalAgentServiceClient, *grpc.ClientConn, error) {
	return NewServiceClient(host, port, apiv1.NewLocalAgentServiceClient)
}

func NewCodebaseAnalysisClient(host string, port int) (apiv1.CodebaseAnalysisServiceClient, *grpc.ClientConn, error) {
	return NewServiceClient(host, port, apiv1.NewCodebaseAnalysisServiceClient)
}

func NewDatabaseClient(host string, port int) (apiv1.DatabaseServiceClient, *grpc.ClientConn, error) {
	return NewServiceClient(host, port, apiv1.NewDatabaseServiceClient)
}

func NewAIClient(host string, port int) (apiv1.AIServiceClient, *grpc.ClientConn, error) {
	return NewServiceClient(host, port, apiv1.NewAIServiceClient)
}

