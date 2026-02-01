package service

import (
	"context"
	"testing"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/services/database/internal/repository"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Getting Started", "getting-started"},
		{"Hello World", "hello-world"},
		{"API Reference", "api-reference"},
		{"README", "readme"},
		{"", ""},
		{"  spaces  ", ""},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.expected {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildSectionTree(t *testing.T) {
	rootID := "root"
	childID := "child"
	sections := []*repository.DocumentSection{
		{ID: "root", Name: "Root", ParentID: nil},
		{ID: "child", Name: "Child", ParentID: &rootID},
		{ID: "other", Name: "Other", ParentID: &childID},
	}
	tree := buildSectionTree(sections, nil)
	if len(tree) != 1 {
		t.Fatalf("expected 1 root section, got %d", len(tree))
	}
	if tree[0].Title != "Root" {
		t.Errorf("root title = %q, want Root", tree[0].Title)
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Title != "Child" {
		t.Errorf("child title = %q, want Child", tree[0].Children[0].Title)
	}
	if len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(tree[0].Children[0].Children))
	}
	if tree[0].Children[0].Children[0].Title != "Other" {
		t.Errorf("grandchild title = %q, want Other", tree[0].Children[0].Children[0].Title)
	}
}

func TestGetDocument_Validation(t *testing.T) {
	svc := NewDatabaseService(repository.NewMockSectionRepository(), repository.NewMockFileItemRepository())
	ctx := context.Background()

	// Empty project_id
	resp, err := svc.GetDocument(ctx, &apiv1.GetDocumentRequest{ProjectId: "", DocumentId: "doc1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrorMsg != "project_id and document_id are required" {
		t.Errorf("expected validation error, got %q", resp.ErrorMsg)
	}

	// Empty document_id
	resp, _ = svc.GetDocument(ctx, &apiv1.GetDocumentRequest{ProjectId: "proj1", DocumentId: ""})
	if resp.ErrorMsg != "project_id and document_id are required" {
		t.Errorf("expected validation error, got %q", resp.ErrorMsg)
	}

	// Not found
	resp, _ = svc.GetDocument(ctx, &apiv1.GetDocumentRequest{ProjectId: "proj1", DocumentId: "nonexistent"})
	if resp.ErrorMsg != "document not found" {
		t.Errorf("expected document not found, got %q", resp.ErrorMsg)
	}
}

func TestGetDocument_Success(t *testing.T) {
	fileRepo := repository.NewMockFileItemRepository()
	fileRepo.UpsertFileItem(context.Background(), &repository.DocumentFileItem{
		ID:        "doc1",
		ProjectID: "proj1",
		Content:   "# Hello World\n\nThis is content.",
		Title:     "Hello",
	})
	svc := NewDatabaseService(repository.NewMockSectionRepository(), fileRepo)
	ctx := context.Background()

	resp, err := svc.GetDocument(ctx, &apiv1.GetDocumentRequest{ProjectId: "proj1", DocumentId: "doc1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrorMsg != "" {
		t.Errorf("unexpected error: %q", resp.ErrorMsg)
	}
	if resp.Doc != "# Hello World\n\nThis is content." {
		t.Errorf("doc = %q, want content", resp.Doc)
	}
}

func TestGetDocumentSections_EmptyProject(t *testing.T) {
	svc := NewDatabaseService(repository.NewMockSectionRepository(), repository.NewMockFileItemRepository())
	ctx := context.Background()

	resp, err := svc.GetDocumentSections(ctx, &apiv1.GetDocumentSectionsRequest{ProjectId: ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Sections) != 0 {
		t.Errorf("expected 0 sections for empty project, got %d", len(resp.Sections))
	}
}

func TestGetDocumentSections_Tree(t *testing.T) {
	sectionRepo := repository.NewMockSectionRepository()
	rootID := "intro"
	ctx := context.Background()
	sectionRepo.UpsertSection(ctx, &repository.DocumentSection{
		ID: "intro", ProjectID: "proj1", Name: "Introduction", URL: "intro", ParentID: nil,
	})
	sectionRepo.UpsertSection(ctx, &repository.DocumentSection{
		ID: "getting-started", ProjectID: "proj1", Name: "Getting Started", URL: "getting-started", ParentID: &rootID,
	})
	sectionRepo.UpsertSection(ctx, &repository.DocumentSection{
		ID: "api", ProjectID: "proj1", Name: "API", URL: "api", ParentID: nil,
	})

	svc := NewDatabaseService(sectionRepo, repository.NewMockFileItemRepository())
	resp, err := svc.GetDocumentSections(ctx, &apiv1.GetDocumentSectionsRequest{ProjectId: "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Sections) != 2 {
		t.Fatalf("expected 2 root sections (Introduction, API), got %d", len(resp.Sections))
	}
	// Find Introduction and check it has Getting Started as child
	var intro *apiv1.DocumentSection
	for _, s := range resp.Sections {
		if s.Title == "Introduction" {
			intro = s
			break
		}
	}
	if intro == nil {
		t.Fatal("Introduction section not found")
	}
	if len(intro.Children) != 1 {
		t.Fatalf("Introduction should have 1 child, got %d", len(intro.Children))
	}
	if intro.Children[0].Title != "Getting Started" {
		t.Errorf("child title = %q, want Getting Started", intro.Children[0].Title)
	}
}

func TestStoreDocument_Validation(t *testing.T) {
	svc := NewDatabaseService(repository.NewMockSectionRepository(), repository.NewMockFileItemRepository())
	ctx := context.Background()

	// Missing id
	resp, _ := svc.StoreDocument(ctx, &apiv1.StoreDocumentRequest{
		Id: "", Title: "Test", ProjectId: "proj1",
	})
	if resp.Success {
		t.Error("expected failure when id is empty")
	}

	// Missing title
	resp, _ = svc.StoreDocument(ctx, &apiv1.StoreDocumentRequest{
		Id: "doc1", Title: "", ProjectId: "proj1",
	})
	if resp.Success {
		t.Error("expected failure when title is empty")
	}

	// Missing project_id
	resp, _ = svc.StoreDocument(ctx, &apiv1.StoreDocumentRequest{
		Id: "doc1", Title: "Test", ProjectId: "",
	})
	if resp.Success {
		t.Error("expected failure when project_id is empty")
	}
}

func TestStoreDocument_Success(t *testing.T) {
	sectionRepo := repository.NewMockSectionRepository()
	fileRepo := repository.NewMockFileItemRepository()
	svc := NewDatabaseService(sectionRepo, fileRepo)
	ctx := context.Background()

	resp, err := svc.StoreDocument(ctx, &apiv1.StoreDocumentRequest{
		Id:          "doc1",
		Title:       "Getting Started",
		Content:     "# Getting Started\n\nWelcome!",
		Description: "Intro section",
		ProjectId:   "proj1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Error("expected success")
	}

	// Verify section was created
	sections, _ := sectionRepo.GetSectionsByProjectID(ctx, "proj1")
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Name != "Getting Started" {
		t.Errorf("section name = %q", sections[0].Name)
	}

	// Verify file item was created
	item, err := fileRepo.GetFileItemByID(ctx, "proj1", "doc1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Content != "# Getting Started\n\nWelcome!" {
		t.Errorf("content = %q", item.Content)
	}
}
