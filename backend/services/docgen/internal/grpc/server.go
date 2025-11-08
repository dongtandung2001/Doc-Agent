package grpc

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/services/docgen/internal/service"
)

type Server struct {
	apiv1.UnimplementedDocumentGenerationServiceServer
	service *service.DocGenService
}

func NewServer(svc *service.DocGenService) *Server {
	return &Server{
		service: svc,
	}
}

func (s *Server) GenerateDocumentSection(
	ctx context.Context,
	req *apiv1.GenerateDocumentSectionRequest,
) (*apiv1.GenerateDocumentSectionResponse, error) {
	return s.service.GenerateDocumentSection(ctx, req)
}

func (s *Server) HealthCheck(
	ctx context.Context,
	req *apiv1.HealthCheckRequest,
) (*apiv1.HealthCheckResponse, error) {
	return &apiv1.HealthCheckResponse{IsAlive: true}, nil
}
