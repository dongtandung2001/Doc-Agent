package repository

import (
	"context"
)

// DocumentSection represents a section in the documentation tree
type DocumentSection struct {
	ID          string
	ProjectID   string
	Description string
	DocumentID  string
	IsCompleted bool
	Name        string
	Order       int
	ParentID    *string
	Prompt      string
	URL         string
}

// DocumentFileItem represents an individual document file
type DocumentFileItem struct {
	ID                string
	ProjectID         string
	Content           string
	Description       string
	DocumentSectionID string
	DocumentID        string
	Extra             string
	IsEmbedded        bool
	Title             string
	EmbedStatus       string // "pending" | "processing" | "completed" | "failed"
}

// SectionRepository defines operations for document sections
type SectionRepository interface {
	// UpsertSection creates or updates a document section
	UpsertSection(ctx context.Context, section *DocumentSection) error
	// GetSectionsByProjectID returns all sections for a project ordered by parent and order
	GetSectionsByProjectID(ctx context.Context, projectID string) ([]*DocumentSection, error)
}

// FileItemRepository defines operations for document file items
type FileItemRepository interface {
	// UpsertFileItem creates or updates a document file item
	UpsertFileItem(ctx context.Context, item *DocumentFileItem) error
	// GetFileItemByID returns a file item by project ID and document ID
	GetFileItemByID(ctx context.Context, projectID, documentID string) (*DocumentFileItem, error)
}
