-- +goose Up

-- Unique indexes for document title uniqueness within same folder
CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_unique_title_root
    ON documents (tenant_id, org_id, lower(title))
    WHERE folder_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_unique_title_folder
    ON documents (tenant_id, org_id, folder_id, lower(title))
    WHERE folder_id IS NOT NULL;

-- Unique indexes for folder name uniqueness within same parent
CREATE UNIQUE INDEX IF NOT EXISTS idx_folders_unique_name_root
    ON folders (tenant_id, org_id, lower(name))
    WHERE parent_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_folders_unique_name_child
    ON folders (tenant_id, org_id, parent_id, lower(name))
    WHERE parent_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_folders_unique_name_child;
DROP INDEX IF EXISTS idx_folders_unique_name_root;
DROP INDEX IF EXISTS idx_documents_unique_title_folder;
DROP INDEX IF EXISTS idx_documents_unique_title_root;
