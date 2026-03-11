DROP INDEX IF EXISTS idx_document_file_items_embed_status;
ALTER TABLE document_file_items DROP COLUMN IF EXISTS embed_status;
