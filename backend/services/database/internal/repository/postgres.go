package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSectionRepository implements SectionRepository for PostgreSQL
type PostgresSectionRepository struct {
	pool *pgxpool.Pool
}

// PostgresFileItemRepository implements FileItemRepository for PostgreSQL
type PostgresFileItemRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSectionRepository creates a new PostgreSQL section repository
func NewPostgresSectionRepository(pool *pgxpool.Pool) *PostgresSectionRepository {
	return &PostgresSectionRepository{pool: pool}
}

// NewPostgresFileItemRepository creates a new PostgreSQL file item repository
func NewPostgresFileItemRepository(pool *pgxpool.Pool) *PostgresFileItemRepository {
	return &PostgresFileItemRepository{pool: pool}
}

// UpsertSection creates or updates a document section, preserving existing fields
// that are not provided in the new section (e.g. parent_id, order, prompt set by StoreSection)
func (r *PostgresSectionRepository) UpsertSection(ctx context.Context, section *DocumentSection) error {
	query := `
		INSERT INTO document_sections (id, project_id, description, document_id, is_completed, name, "order", parent_id, prompt, url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (id) DO UPDATE SET
			project_id = EXCLUDED.project_id,
			description = CASE WHEN EXCLUDED.description != '' THEN EXCLUDED.description ELSE document_sections.description END,
			document_id = EXCLUDED.document_id,
			is_completed = EXCLUDED.is_completed,
			name = EXCLUDED.name,
			"order" = CASE WHEN EXCLUDED."order" != 0 THEN EXCLUDED."order" ELSE document_sections."order" END,
			parent_id = CASE WHEN EXCLUDED.parent_id IS NOT NULL THEN EXCLUDED.parent_id ELSE document_sections.parent_id END,
			prompt = CASE WHEN EXCLUDED.prompt != '' THEN EXCLUDED.prompt ELSE document_sections.prompt END,
			url = EXCLUDED.url,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		section.ID,
		section.ProjectID,
		section.Description,
		section.DocumentID,
		section.IsCompleted,
		section.Name,
		section.Order,
		section.ParentID,
		section.Prompt,
		section.URL,
	)
	return err
}

// GetSectionsByProjectID returns all sections for a project
func (r *PostgresSectionRepository) GetSectionsByProjectID(ctx context.Context, projectID string) ([]*DocumentSection, error) {
	query := `
		SELECT id, project_id, description, document_id, is_completed, name, "order", parent_id, prompt, url
		FROM document_sections
		WHERE project_id = $1
		ORDER BY COALESCE(parent_id, ''), "order", name
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []*DocumentSection
	for rows.Next() {
		var s DocumentSection
		var parentID *string
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Description, &s.DocumentID, &s.IsCompleted, &s.Name, &s.Order, &parentID, &s.Prompt, &s.URL); err != nil {
			return nil, err
		}
		s.ParentID = parentID
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}

// UpsertFileItem creates or updates a document file item.
// embed_status is always reset to 'pending' so the RAG pipeline picks it up.
func (r *PostgresFileItemRepository) UpsertFileItem(ctx context.Context, item *DocumentFileItem) error {
	query := `
		INSERT INTO document_file_items (id, project_id, content, description, document_section_id, document_id, extra, is_embedded, title, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (id) DO UPDATE SET
			project_id          = EXCLUDED.project_id,
			content             = EXCLUDED.content,
			description         = EXCLUDED.description,
			document_section_id = EXCLUDED.document_section_id,
			document_id         = EXCLUDED.document_id,
			extra               = EXCLUDED.extra,
			is_embedded         = EXCLUDED.is_embedded,
			title               = EXCLUDED.title,
			embed_status        = 'pending',
			updated_at          = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		item.ID,
		item.ProjectID,
		item.Content,
		item.Description,
		item.DocumentSectionID,
		item.DocumentID,
		item.Extra,
		item.IsEmbedded,
		item.Title,
	)
	return err
}

// ErrNotFound is returned when a document is not found
var ErrNotFound = errors.New("document not found")

// GetFileItemByID returns a file item by project ID and document ID
func (r *PostgresFileItemRepository) GetFileItemByID(ctx context.Context, projectID, documentID string) (*DocumentFileItem, error) {
	query := `
		SELECT id, project_id, content, description, document_section_id, document_id, extra, is_embedded, title, embed_status
		FROM document_file_items
		WHERE project_id = $1 AND document_id = $2
	`
	var item DocumentFileItem
	err := r.pool.QueryRow(ctx, query, projectID, documentID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Content,
		&item.Description,
		&item.DocumentSectionID,
		&item.DocumentID,
		&item.Extra,
		&item.IsEmbedded,
		&item.Title,
		&item.EmbedStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}
