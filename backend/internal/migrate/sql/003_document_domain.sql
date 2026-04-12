-- +goose Up

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_status enum
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE document_status AS ENUM ('pending', 'processing', 'processed', 'failed');

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_type enum
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE document_type AS ENUM (
    'invoice', 'contract', 'identity', 'warranty', 'receipt', 'other'
);

-- ═══════════════════════════════════════════════════════════════════════════════
-- folders
-- Nested folder tree. parent_id is self-referential.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE folders (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES folders(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_folders_tenant ON folders(tenant_id);
CREATE INDEX idx_folders_org ON folders(org_id);
CREATE INDEX idx_folders_parent ON folders(parent_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- documents
-- Stable identity record for each uploaded document.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE documents (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    folder_id   UUID REFERENCES folders(id) ON DELETE SET NULL,
    owner_id    UUID NOT NULL REFERENCES users(id),
    title       TEXT NOT NULL,
    doc_type    document_type NOT NULL DEFAULT 'other',
    status      document_status NOT NULL DEFAULT 'pending',
    language    TEXT,  -- detected language of the document
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_tenant ON documents(tenant_id);
CREATE INDEX idx_documents_org ON documents(org_id);
CREATE INDEX idx_documents_folder ON documents(folder_id);
CREATE INDEX idx_documents_owner ON documents(owner_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_type ON documents(doc_type);
CREATE INDEX idx_documents_created ON documents(created_at DESC);

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_versions
-- Each upload or replace creates a new version. Original never deleted.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE document_versions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id     UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_number  INTEGER NOT NULL,
    storage_key     TEXT NOT NULL,  -- MinIO object key
    mime_type       TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    uploaded_by     UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(document_id, version_number)
);

CREATE INDEX idx_doc_versions_document ON document_versions(document_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_pages
-- Per-page OCR output. confidence is float 0-1.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE document_pages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id     UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_id      UUID NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    page_number     INTEGER NOT NULL,
    ocr_text        TEXT,
    confidence      REAL CHECK (confidence >= 0 AND confidence <= 1),
    ocr_model       TEXT NOT NULL DEFAULT 'mistral-ocr-2503',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(document_id, version_id, page_number)
);

CREATE INDEX idx_doc_pages_document ON document_pages(document_id);
CREATE INDEX idx_doc_pages_version ON document_pages(version_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_metadata
-- Key-value metadata. Extracted and corrected values stored separately.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE document_metadata (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id       UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    key               TEXT NOT NULL,   -- issuer, amount, currency, issue_date, expiry_date, document_number, language
    extracted_value   TEXT,            -- value extracted by OCR/ML pipeline
    corrected_value   TEXT,            -- user-corrected value (takes precedence)
    corrected_by      UUID REFERENCES users(id),
    corrected_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(document_id, key)
);

CREATE INDEX idx_doc_metadata_document ON document_metadata(document_id);
CREATE INDEX idx_doc_metadata_key ON document_metadata(key);

-- ═══════════════════════════════════════════════════════════════════════════════
-- extracted_text_chunks
-- Retrieval unit with pgvector embedding for semantic search.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE extracted_text_chunks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id     UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_id         UUID NOT NULL REFERENCES document_pages(id) ON DELETE CASCADE,
    chunk_index     INTEGER NOT NULL,
    chunk_text      TEXT NOT NULL,
    embedding       vector(1024),  -- pgvector column (mistral-embed-2312 = 1024 dimensions)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(document_id, page_id, chunk_index)
);

CREATE INDEX idx_chunks_document ON extracted_text_chunks(document_id);
CREATE INDEX idx_chunks_page ON extracted_text_chunks(page_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- tags
-- Global tag registry per tenant.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE tags (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_tags_tenant ON tags(tenant_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- document_tags
-- Many-to-many join between documents and tags.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE document_tags (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (document_id, tag_id)
);

CREATE INDEX idx_doc_tags_tag ON document_tags(tag_id);

-- +goose Down
DROP TABLE IF EXISTS document_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS extracted_text_chunks;
DROP TABLE IF EXISTS document_metadata;
DROP TABLE IF EXISTS document_pages;
DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS folders;
DROP TYPE IF EXISTS document_type;
DROP TYPE IF EXISTS document_status;
