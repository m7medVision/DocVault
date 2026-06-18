-- +goose Up
ALTER TABLE folders ADD COLUMN index_content TEXT;

-- +goose Down
ALTER TABLE folders DROP COLUMN IF EXISTS index_content;
