-- DocumentSections: represents folders or categories of documents
CREATE TABLE IF NOT EXISTS document_sections (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    description TEXT,
    document_id VARCHAR(255),
    is_completed BOOLEAN DEFAULT FALSE,
    name VARCHAR(500) NOT NULL,
    "order" INTEGER DEFAULT 0,
    parent_id VARCHAR(255),
    prompt TEXT,
    url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- DocumentFileItems: represents individual files or document entries
CREATE TABLE IF NOT EXISTS document_file_items (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    content TEXT,
    description TEXT,
    document_section_id VARCHAR(255) NOT NULL REFERENCES document_sections(id) ON DELETE CASCADE,
    document_id VARCHAR(255),
    extra TEXT,
    is_embedded BOOLEAN DEFAULT TRUE,
    title VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_document_sections_project_id ON document_sections(project_id);
CREATE INDEX IF NOT EXISTS idx_document_sections_parent_id ON document_sections(parent_id);
CREATE INDEX IF NOT EXISTS idx_document_sections_project_order ON document_sections(project_id, "order");
CREATE INDEX IF NOT EXISTS idx_document_file_items_project_id ON document_file_items(project_id);
CREATE INDEX IF NOT EXISTS idx_document_file_items_document_section_id ON document_file_items(document_section_id);
CREATE INDEX IF NOT EXISTS idx_document_file_items_project_document ON document_file_items(project_id, id);
