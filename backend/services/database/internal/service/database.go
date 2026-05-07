package service

import (
	"context"
	"errors"
	"strings"

	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/repository"
	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
)

type DatabaseService struct {
	sectionRepo repository.SectionRepository
	fileRepo    repository.FileItemRepository
}

func NewDatabaseService(
	sectionRepo repository.SectionRepository,
	fileRepo repository.FileItemRepository,
) *DatabaseService {
	return &DatabaseService{
		sectionRepo: sectionRepo,
		fileRepo:    fileRepo,
	}
}

func (s *DatabaseService) GetDocument(
	ctx context.Context,
	req *apiv1.GetDocumentRequest,
) (*apiv1.GetDocumentResponse, error) {
	if req.ProjectId == "" || req.DocumentId == "" {
		return &apiv1.GetDocumentResponse{
			ErrorMsg: "project_id and document_id are required",
		}, nil
	}

	item, err := s.fileRepo.GetFileItemByID(ctx, req.ProjectId, req.DocumentId)
	if err != nil {
		errMsg := err.Error()
		if errors.Is(err, repository.ErrNotFound) {
			errMsg = "document not found"
		}
		return &apiv1.GetDocumentResponse{
			ErrorMsg: errMsg,
		}, nil
	}

	return &apiv1.GetDocumentResponse{
		Doc: item.Content,
	}, nil
}

func (s *DatabaseService) GetDocumentSections(
	ctx context.Context,
	req *apiv1.GetDocumentSectionsRequest,
) (*apiv1.GetDocumentSectionsResponse, error) {
	if req.ProjectId == "" {
		return &apiv1.GetDocumentSectionsResponse{
			Sections: []*apiv1.DocumentSection{},
		}, nil
	}

	sections, err := s.sectionRepo.GetSectionsByProjectID(ctx, req.ProjectId)
	if err != nil {
		return &apiv1.GetDocumentSectionsResponse{
			Sections: []*apiv1.DocumentSection{},
		}, err
	}

	// Build nested tree from flat list
	tree := buildSectionTree(sections, nil)
	return &apiv1.GetDocumentSectionsResponse{
		Sections: tree,
	}, nil
}

func (s *DatabaseService) StoreDocument(
	ctx context.Context,
	req *apiv1.StoreDocumentRequest,
) (*apiv1.StoreDocumentResponse, error) {
	// Validate required fields
	if req.Id == "" || req.Title == "" {
		return &apiv1.StoreDocumentResponse{Success: false}, nil
	}

	// document_id is the section this file belongs to (or creates)
	sectionID := req.DocumentId
	if sectionID == "" {
		// Use a slug from title as section id if not provided
		sectionID = slugify(req.Title)
	}

	// project_id is required for scoping - use from request or derive
	projectID := req.ProjectId
	if projectID == "" {
		return &apiv1.StoreDocumentResponse{Success: false}, nil
	}

	// Upsert the section first (so the FK constraint is satisfied)
	section := &repository.DocumentSection{
		ID:          sectionID,
		ProjectID:   projectID,
		Name:        req.Title,
		Description: req.Description,
		URL:         slugify(req.Title),
		Order:       0,
	}
	if err := s.sectionRepo.UpsertSection(ctx, section); err != nil {
		return &apiv1.StoreDocumentResponse{Success: false}, err
	}

	// Upsert the file item
	item := &repository.DocumentFileItem{
		ID:                req.Id,
		ProjectID:         projectID,
		Content:           req.Content,
		Description:       req.Description,
		DocumentSectionID: sectionID,
		DocumentID:        sectionID,
		Extra:             req.Extra,
		IsEmbedded:        req.IsEmbedded,
		Title:             req.Title,
	}
	if err := s.fileRepo.UpsertFileItem(ctx, item); err != nil {
		return &apiv1.StoreDocumentResponse{Success: false}, err
	}

	return &apiv1.StoreDocumentResponse{Success: true}, nil
}

func (s *DatabaseService) StoreSection(
	ctx context.Context,
	req *apiv1.StoreSectionRequest,
) (*apiv1.StoreSectionResponse, error) {
	if req.Id == "" || req.Title == "" || req.ProjectId == "" {
		return &apiv1.StoreSectionResponse{Success: false}, nil
	}

	url := req.Url
	if url == "" {
		url = req.Id
	}

	var parentID *string
	if req.ParentId != "" {
		parentID = &req.ParentId
	}

	section := &repository.DocumentSection{
		ID:        req.Id,
		ProjectID: req.ProjectId,
		Name:      req.Title,
		Prompt:    req.Prompt,
		URL:       url,
		Order:     int(req.Order),
		ParentID:  parentID,
	}
	if err := s.sectionRepo.UpsertSection(ctx, section); err != nil {
		return &apiv1.StoreSectionResponse{Success: false}, err
	}

	return &apiv1.StoreSectionResponse{Success: true}, nil
}

// buildSectionTree converts flat sections into nested tree structure
func buildSectionTree(sections []*repository.DocumentSection, parentID *string) []*apiv1.DocumentSection {
	var result []*apiv1.DocumentSection
	for _, s := range sections {
		// Match by parent: nil parentID means root level (sections with no parent)
		// Otherwise match sections whose ParentID equals parentID
		var isRoot, isChild bool
		if parentID == nil {
			isRoot = s.ParentID == nil || *s.ParentID == ""
		} else {
			isChild = s.ParentID != nil && *s.ParentID == *parentID
		}
		if isRoot || isChild {
			section := &apiv1.DocumentSection{
				Title:       s.Name,
				Description: s.Description,
				Url:         s.DocumentID,
				Children:    buildSectionTree(sections, ptr(s.ID)),
			}
			result = append(result, section)
		}
	}
	return result
}

func ptr(s string) *string { return &s }

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}
