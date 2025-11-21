package grpc

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/services/codebase/internal/service"
)

// Server implements the gRPC CodebaseAnalysisServiceServer interface
type Server struct {
	apiv1.UnimplementedCodebaseAnalysisServiceServer
	service *service.AnalysisService
}

func NewServer(svc *service.AnalysisService) *Server {
	return &Server{
		service: svc,
	}
}

// StartCodebaseAnalysis handles the codebase analysis request
func (s *Server) StartCodebaseAnalysis(
	ctx context.Context,
	req *apiv1.StartCodebaseAnalysisRequest,
) (*apiv1.StartCodebaseAnalysisResponse, error) {
	return s.service.StartCodebaseAnalysis(ctx, req)
}

// HealthCheck returns the health status of the service
func (s *Server) HealthCheck(
	ctx context.Context,
	req *apiv1.HealthCheckRequest,
) (*apiv1.HealthCheckResponse, error) {
	// Check if service is healthy (can add checks for dependencies, DB, etc.)
	return &apiv1.HealthCheckResponse{IsAlive: true}, nil
}