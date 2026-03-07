package handlers

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
)

// AIClient is the interface used by the gateway to call the AI (RAG/Chat) service.
// It can be implemented by the gRPC client or by HTTPAIClient (Python proxy).
type AIClient interface {
	Chat(ctx context.Context, req *apiv1.ChatRequest) (*apiv1.ChatResponse, error)
	HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error)
}

// GatewayHandler implements GatewayService and proxies requests to backend services
type GatewayHandler struct {
	localAgentClient apiv1.LocalAgentServiceClient
	codebaseClient   apiv1.CodebaseAnalysisServiceClient
	databaseClient   apiv1.DatabaseServiceClient
	aiClient         AIClient
}

func NewGatewayHandler(
	localAgentClient apiv1.LocalAgentServiceClient,
	codebaseClient apiv1.CodebaseAnalysisServiceClient,
	databaseClient apiv1.DatabaseServiceClient,
	aiClient AIClient,
) *GatewayHandler {
	return &GatewayHandler{
		localAgentClient: localAgentClient,
		codebaseClient:   codebaseClient,
		databaseClient:   databaseClient,
		aiClient:         aiClient,
	}
}

// StartCodebaseAnalysis proxies to Codebase Analysis Service
func (h *GatewayHandler) StartCodebaseAnalysis(
	ctx context.Context,
	req *connect.Request[apiv1.StartCodebaseAnalysisRequest],
) (*connect.Response[apiv1.StartCodebaseAnalysisResponse], error) {
	result, err := h.codebaseClient.StartCodebaseAnalysis(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

// RequestFileContent proxies to Local Agent Service
func (h *GatewayHandler) RequestFileContent(
	ctx context.Context,
	req *connect.Request[apiv1.RequestFileContentRequest],
) (*connect.Response[apiv1.RequestFileContentResponse], error) {
	result, err := h.localAgentClient.RequestFileContent(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

// GetDocument proxies to Database Service
func (h *GatewayHandler) GetDocument(
	ctx context.Context,
	req *connect.Request[apiv1.GetDocumentRequest],
) (*connect.Response[apiv1.GetDocumentResponse], error) {
	result, err := h.databaseClient.GetDocument(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

func (h *GatewayHandler) StoreDocument(
	ctx context.Context,
	req *connect.Request[apiv1.StoreDocumentRequest],
) (*connect.Response[apiv1.StoreDocumentResponse], error) {
	result, err := h.databaseClient.StoreDocument(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

// GetDocumentSections proxies to Database Service
func (h *GatewayHandler) GetDocumentSections(
	ctx context.Context,
	req *connect.Request[apiv1.GetDocumentSectionsRequest],
) (*connect.Response[apiv1.GetDocumentSectionsResponse], error) {
	result, err := h.databaseClient.GetDocumentSections(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

// Chat proxies to AI Service
func (h *GatewayHandler) Chat(
	ctx context.Context,
	req *connect.Request[apiv1.ChatRequest],
) (*connect.Response[apiv1.ChatResponse], error) {
	if h.aiClient == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("AI service not configured: set backends.ai.base_url in gateway config"))
	}
	result, err := h.aiClient.Chat(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(result), nil
}

// HealthCheck checks health of all backend services
func (h *GatewayHandler) HealthCheck(
	ctx context.Context,
	req *connect.Request[apiv1.HealthCheckRequest],
) (*connect.Response[apiv1.HealthCheckResponse], error) {
	// Check all services - skip nil clients (for testing with subset of services)

	// Check Local Agent (skip if nil)
	if h.localAgentClient != nil {
		if _, err := h.localAgentClient.HealthCheck(ctx, req.Msg); err != nil {
			log.Printf("Local Agent HealthCheck error: %v", err)
			return connect.NewResponse(&apiv1.HealthCheckResponse{IsAlive: false}), nil
		}
	}

	// Check Codebase Analysis (skip if nil)
	if h.codebaseClient != nil {
		if _, err := h.codebaseClient.HealthCheck(ctx, req.Msg); err != nil {
			return connect.NewResponse(&apiv1.HealthCheckResponse{IsAlive: false}), nil
		}
	}

	// Check Database Service (skip if nil)
	if h.databaseClient != nil {
		if _, err := h.databaseClient.HealthCheck(ctx, req.Msg); err != nil {
			return connect.NewResponse(&apiv1.HealthCheckResponse{IsAlive: false}), nil
		}
	}

	// Check AI Service (skip if nil)
	if h.aiClient != nil {
		if _, err := h.aiClient.HealthCheck(ctx, req.Msg); err != nil {
			return connect.NewResponse(&apiv1.HealthCheckResponse{IsAlive: false}), nil
		}
	}

	// All non-nil services are healthy
	return connect.NewResponse(&apiv1.HealthCheckResponse{IsAlive: true}), nil
}
