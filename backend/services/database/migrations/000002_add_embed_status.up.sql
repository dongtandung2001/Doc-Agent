-- Add embed_status to track vector-embedding pipeline state independently from is_embedded.
-- is_embedded means "content is stored in the DB"; embed_status tracks whether that content
-- has been processed by the RAG vector pipeline.
--
-- States:
--   pending    - content exists, not yet picked up by the embedding worker
--   processing - currently being embedded (worker lock)
--   completed  - successfully stored in the vector store
--   failed     - embedding attempt failed (will not be retried automatically)

ALTER TABLE document_file_items
    ADD COLUMN IF NOT EXISTS embed_status VARCHAR(20) NOT NULL DEFAULT 'pending';

-- Any row that already has content but was never tracked should start as pending.
-- Rows with empty content are skipped by the pipeline anyway.
CREATE INDEX IF NOT EXISTS idx_document_file_items_embed_status
    ON document_file_items(project_id, embed_status)
    WHERE content IS NOT NULL AND content != '';
