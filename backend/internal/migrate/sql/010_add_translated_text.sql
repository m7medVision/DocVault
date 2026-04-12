-- +goose Up

ALTER TABLE document_pages
    ADD COLUMN IF NOT EXISTS translated_text TEXT;

-- +goose Down

ALTER TABLE document_pages
    DROP COLUMN IF EXISTS translated_text;
