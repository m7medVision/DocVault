-- +goose Up

-- Processing stage tracking + AI suggestion output persistence.
-- Stage transitions: uploaded → ocr_queued → ocr_running → ocr_complete → processing_queued → processing_running → indexing → suggesting → completed
-- Terminal failure states: ocr_failed, processing_failed

ALTER TABLE documents ADD COLUMN processing_stage VARCHAR(32) DEFAULT NULL;
ALTER TABLE documents ADD COLUMN processing_error TEXT DEFAULT NULL;
ALTER TABLE documents ADD COLUMN suggested_folder_name TEXT DEFAULT NULL;
ALTER TABLE documents ADD COLUMN suggested_filename TEXT DEFAULT NULL;
ALTER TABLE documents ADD COLUMN suggestion_confidence REAL DEFAULT NULL;
ALTER TABLE documents ADD COLUMN suggestion_create_new BOOLEAN DEFAULT NULL;

CREATE INDEX idx_documents_processing_stage ON documents(processing_stage);
CREATE INDEX idx_documents_processing_error ON documents(processing_error) WHERE processing_error IS NOT NULL;

-- +goose Down

ALTER TABLE documents DROP COLUMN IF EXISTS processing_stage;
ALTER TABLE documents DROP COLUMN IF EXISTS processing_error;
ALTER TABLE documents DROP COLUMN IF EXISTS suggested_folder_name;
ALTER TABLE documents DROP COLUMN IF EXISTS suggested_filename;
ALTER TABLE documents DROP COLUMN IF EXISTS suggestion_confidence;
ALTER TABLE documents DROP COLUMN IF EXISTS suggestion_create_new;
