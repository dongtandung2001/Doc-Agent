package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
)

// HTTPAIClient forwards Chat and HealthCheck to a Python AI service HTTP wrapper
// (e.g. serve_web_test.py) that proxies to the gRPC AI service.
// This allows the gateway to use the Python RAG/Chat service without proto alignment.
type HTTPAIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPAIClient creates a client that POSTs to baseURL (e.g. "http://localhost:8888").
func NewHTTPAIClient(baseURL string) *HTTPAIClient {
	return &HTTPAIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends messages and project_id to the AI HTTP API and returns content.
func (c *HTTPAIClient) Chat(ctx context.Context, req *apiv1.ChatRequest) (*apiv1.ChatResponse, error) {
	body := struct {
		Messages  []*apiv1.ChatMessage `json:"messages"`
		ProjectID string               `json:"project_id"`
	}{
		Messages:  req.Messages,
		ProjectID: req.ProjectId,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat API returned %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode chat response: %w", err)
	}

	return &apiv1.ChatResponse{Content: result.Content}, nil
}

// HealthCheck calls the AI HTTP health endpoint.
func (c *HTTPAIClient) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/health", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &apiv1.HealthCheckResponse{IsAlive: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &apiv1.HealthCheckResponse{IsAlive: false}, nil
	}

	var result struct {
		IsAlive bool `json:"isAlive"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return &apiv1.HealthCheckResponse{IsAlive: result.IsAlive}, nil
}
