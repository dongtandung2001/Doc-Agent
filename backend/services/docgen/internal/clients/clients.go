package clients

import (
	"fmt"
	"net"
	"strconv"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAIClient(host string, port int) (apiv1.AIServiceClient, *grpc.ClientConn, error) {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AI service client: %w", err)
	}

	return apiv1.NewAIServiceClient(conn), conn, nil
}

func NewGatewayClient(host string, port int) (apiv1.GatewayServiceClient, *grpc.ClientConn, error) {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Gateway client: %w", err)
	}

	return apiv1.NewGatewayServiceClient(conn), conn, nil
}
