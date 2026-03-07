package grpc

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/service"
)

type Server struct {
	apiv1.UnimplementedDatabaseServiceServer
	service *service.DatabaseService
}

func NewServer(svc *service.DatabaseService) *Server {
	return &Server{
		service: svc,
	}
}

func (s *Server) GetDocument(
	ctx context.Context,
	req *apiv1.GetDocumentRequest,
) (*apiv1.GetDocumentResponse, error) {
	return s.service.GetDocument(ctx, req)
}

func (s *Server) GetDocumentSections(
	ctx context.Context,
	req *apiv1.GetDocumentSectionsRequest,
) (*apiv1.GetDocumentSectionsResponse, error) {
	return s.service.GetDocumentSections(ctx, req)
}

func (s *Server) StoreDocument(
	ctx context.Context,
	req *apiv1.StoreDocumentRequest,
) (*apiv1.StoreDocumentResponse, error) {
	return s.service.StoreDocument(ctx, req)
}

func (s *Server) StoreSection(
	ctx context.Context,
	req *apiv1.StoreSectionRequest,
) (*apiv1.StoreSectionResponse, error) {
	return s.service.StoreSection(ctx, req)
}

func (s *Server) HealthCheck(
	ctx context.Context,
	req *apiv1.HealthCheckRequest,
) (*apiv1.HealthCheckResponse, error) {
	return &apiv1.HealthCheckResponse{IsAlive: true}, nil
}
