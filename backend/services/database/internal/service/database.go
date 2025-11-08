package service

import (
	"context"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
)

type DatabaseService struct {
	// Add dependencies: SQL DB client, Vector DB client
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

func (s *DatabaseService) GetDocument(
	ctx context.Context,
	req *apiv1.GetDocumentRequest,
) (*apiv1.GetDocumentResponse, error) {
	// TODO: Implement document retrieval
	return &apiv1.GetDocumentResponse{}, nil
}

func (s *DatabaseService) GetDocumentSections(
	ctx context.Context,
	req *apiv1.GetDocumentSectionsRequest,
) (*apiv1.GetDocumentSectionsResponse, error) {
	// TODO: Implement section retrieval
	return &apiv1.GetDocumentSectionsResponse{}, nil
}

func (s *DatabaseService) StoreDocument(
	ctx context.Context,
	req *apiv1.StoreDocumentRequest,
) (*apiv1.StoreDocumentResponse, error) {
	// TODO: Implement document storing
	return &apiv1.StoreDocumentResponse{Success: true}, nil
}
