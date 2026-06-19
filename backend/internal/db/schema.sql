-- Source: internal/migrate/sql/001_enable_extensions.sql
-- Enable pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable citext for case-insensitive text
CREATE EXTENSION IF NOT EXISTS citext;

-- Enable pg_trgm for fuzzy keyword / proper-noun similarity (word_similarity)
CREATE EXTENSION IF NOT EXISTS pg_trgm;


-- Source: internal/migrate/sql/002_identity_access.sql

-- ═══════════════════════════════════════════════════════════════════════════════
-- tenants
-- Root isolation unit. All other tables reference tenant_id.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    plan        TEXT NOT NULL DEFAULT 'free',  -- free | personal | business
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tenants_plan ON tenants(plan);

-- ═══════════════════════════════════════════════════════════════════════════════
-- organizations
-- Sub-unit within a tenant for multi-org setups.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orgs_tenant ON organizations(tenant_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- users
-- Internal user identities scoped to a tenant.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email           CITEXT NOT NULL,
    display_name    TEXT NOT NULL,
    locale          TEXT NOT NULL DEFAULT 'en' CHECK (locale IN ('ar', 'en')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

-- ═══════════════════════════════════════════════════════════════════════════════
-- memberships
-- Links users to organizations with a role.
-- Roles: owner, admin, member, viewer
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE membership_role AS ENUM ('owner', 'admin', 'member', 'viewer');

CREATE TABLE memberships (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role        membership_role NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Each user can only have one membership per org
    UNIQUE(user_id, org_id)
);

CREATE INDEX idx_memberships_user ON memberships(user_id);
CREATE INDEX idx_memberships_org ON memberships(org_id);
CREATE INDEX idx_memberships_role ON memberships(role);


-- Source: internal/migrate/sql/003_document_domain.sql

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
    -- Maintained by Postgres from chunk_text; powers keyword / phrase retrieval.
    -- 'simple' config is language-agnostic, so it works for AR, EN, proper nouns,
    -- IDs and dates alike (no stemming, every token indexed verbatim).
    chunk_tsv       tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(chunk_text, ''))) STORED,
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


-- Source: internal/migrate/sql/004_reminder_audit.sql

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_rule_source enum
-- auto = derived by Reminder Worker from OCR
-- manual = user-created
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE reminder_rule_source AS ENUM ('auto', 'manual');

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_rules
-- Rules that define when to send reminders for a document.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE reminder_rules (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id           UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id             UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rule_type             TEXT NOT NULL,              -- expiry, renewal, due_date
    trigger_date          DATE NOT NULL,              -- the date that triggers the reminder
    notify_days_before    INTEGER[] NOT NULL DEFAULT '{30, 7, 1, 0}',  -- days before trigger_date to notify
    source                reminder_rule_source NOT NULL DEFAULT 'auto',
    active                BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reminder_rules_document ON reminder_rules(document_id);
CREATE INDEX idx_reminder_rules_tenant ON reminder_rules(tenant_id);
CREATE INDEX idx_reminder_rules_trigger ON reminder_rules(trigger_date);
CREATE INDEX idx_reminder_rules_active ON reminder_rules(active) WHERE active = true;

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_event_status enum
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE reminder_event_status AS ENUM ('pending', 'sent', 'failed', 'snoozed');

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_events
-- Tracks delivery state per notification.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE reminder_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id         UUID NOT NULL REFERENCES reminder_rules(id) ON DELETE CASCADE,
    scheduled_at    TIMESTAMPTZ NOT NULL,              -- when the notification should be sent
    sent_at         TIMESTAMPTZ,                       -- when it was actually sent
    channel         TEXT NOT NULL DEFAULT 'in_app',    -- in_app, email
    status          reminder_event_status NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reminder_events_rule ON reminder_events(rule_id);
CREATE INDEX idx_reminder_events_status ON reminder_events(status);
CREATE INDEX idx_reminder_events_scheduled ON reminder_events(scheduled_at) WHERE status = 'pending';

-- ═══════════════════════════════════════════════════════════════════════════════
-- audit_events
-- Append-only audit log. Cannot be modified by any user role.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE audit_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    actor_id    UUID REFERENCES users(id),             -- null for system actions
    entity_type TEXT NOT NULL,                          -- document, user, folder, reminder
    entity_id   UUID NOT NULL,
    action      TEXT NOT NULL,                          -- upload, download, delete, metadata.correct, member.invite, etc.
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_tenant ON audit_events(tenant_id);
CREATE INDEX idx_audit_actor ON audit_events(actor_id);
CREATE INDEX idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_action ON audit_events(action);
CREATE INDEX idx_audit_created ON audit_events(created_at DESC);

-- Prevent updates and deletes on audit_events (enforce append-only)
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only; updates and deletes are not permitted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_events_no_update
    BEFORE UPDATE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();

CREATE TRIGGER audit_events_no_delete
    BEFORE DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();


-- Source: internal/migrate/sql/005_pgvector_index.sql

-- HNSW index for approximate nearest-neighbour cosine similarity search
-- m=16: number of bi-directional links per node (good balance for 1024-dim vectors)
-- ef_construction=64: search width during index build (higher = better recall, slower build)
CREATE INDEX IF NOT EXISTS idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- GIN index over the generated chunk_tsv column for keyword / phrase matching
-- (to_tsquery @@ chunk_tsv).
CREATE INDEX IF NOT EXISTS idx_chunks_fts
    ON extracted_text_chunks
    USING gin (chunk_tsv);

-- Trigram GIN indexes backing word_similarity() fuzzy containment for proper
-- nouns, IDs, dates and typos (e.g. "teepee" vs "tepee").
CREATE INDEX IF NOT EXISTS idx_chunks_trgm
    ON extracted_text_chunks
    USING gin (chunk_text gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_documents_title_trgm
    ON documents
    USING gin (title gin_trgm_ops);


-- Source: internal/migrate/sql/006_add_user_auth_fields.sql

-- Add password hash field (bcrypt hashes are max 60 chars, but we use 255 for future-proofing)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Add email verified flag
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE NOT NULL;

-- Add last login timestamp
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Add failed login attempts counter (for brute force protection)
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0 NOT NULL;

-- Add account locked timestamp (for lockout mechanism)
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;

-- Add index on email for fast lookup during login
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add index on tenant_id + email for multi-tenant queries
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);

-- Add comment explaining the password hash field
COMMENT ON COLUMN users.password_hash IS 'bcrypt hash of user password (cost factor 12)';
COMMENT ON COLUMN users.email_verified IS 'Whether user has verified their email address';
COMMENT ON COLUMN users.last_login_at IS 'Timestamp of last successful login';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts (reset on successful login)';
COMMENT ON COLUMN users.locked_until IS 'Timestamp until which account is locked due to too many failed attempts';


-- Source: internal/migrate/sql/007_create_casbin_rule.sql

CREATE TABLE IF NOT EXISTS casbin_rule (
    id      BIGSERIAL PRIMARY KEY,
    ptype   TEXT NOT NULL,
    v0      TEXT NOT NULL DEFAULT '',
    v1      TEXT NOT NULL DEFAULT '',
    v2      TEXT NOT NULL DEFAULT '',
    v3      TEXT NOT NULL DEFAULT '',
    v4      TEXT NOT NULL DEFAULT '',
    v5      TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_casbin_rule_unique
    ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

CREATE INDEX IF NOT EXISTS idx_casbin_rule_ptype ON casbin_rule (ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_rule_domain ON casbin_rule (v1);

COMMENT ON TABLE casbin_rule IS 'Casbin RBAC policy table';
COMMENT ON COLUMN casbin_rule.v0 IS 'Subject or user/role binding source';
COMMENT ON COLUMN casbin_rule.v1 IS 'Tenant/domain';
COMMENT ON COLUMN casbin_rule.v2 IS 'Object or role binding target/domain field';
COMMENT ON COLUMN casbin_rule.v3 IS 'Action';


-- Source: internal/migrate/sql/008_create_notifications.sql

-- ═══════════════════════════════════════════════════════════════════════════════
-- notifications
-- In-app notification records for users.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT NOT NULL DEFAULT 'reminder',
    title       TEXT NOT NULL,
    body        TEXT,
    link        TEXT,
    status      TEXT NOT NULL DEFAULT 'unread',
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at     TIMESTAMPTZ
);

CREATE INDEX idx_notifications_tenant_user ON notifications (tenant_id, user_id);
CREATE INDEX idx_notifications_status ON notifications (status);
CREATE INDEX idx_notifications_created ON notifications (created_at DESC);


-- Source: internal/migrate/sql/009_embedding_1024.sql

-- Recreate the embedding column and HNSW index for mistral-embed-2312 (1024 dimensions).
-- Must drop the index first, then alter the column type, then rebuild the index.

DROP INDEX IF EXISTS idx_chunks_embedding_hnsw;

ALTER TABLE extracted_text_chunks
    ALTER COLUMN embedding TYPE vector(1024);

-- Rebuild HNSW index tuned for 1024-dim vectors
-- m=16: bi-directional links per node (good balance)
-- ef_construction=64: search width during build (higher = better recall)
CREATE INDEX idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);


-- Source: internal/migrate/sql/010_add_translated_text.sql

ALTER TABLE document_pages
    ADD COLUMN IF NOT EXISTS translated_text TEXT;


-- Source: internal/migrate/sql/011_add_processing_stage_and_suggestions.sql

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


-- Source: internal/migrate/sql/012_auto_organize_cleanup.sql

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


-- Source: internal/migrate/sql/013_acl_and_restrictions.sql
CREATE TYPE acl_resource_type  AS ENUM ('document', 'folder');
CREATE TYPE acl_principal_type AS ENUM ('user', 'group');
CREATE TYPE acl_permission     AS ENUM ('read', 'write', 'delete');

CREATE TABLE groups (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, name)
);
CREATE INDEX idx_groups_tenant ON groups(tenant_id);
CREATE INDEX idx_groups_org    ON groups(org_id);

CREATE TABLE group_members (
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);
CREATE INDEX idx_group_members_user ON group_members(user_id);

CREATE TABLE acl_grants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    resource_type   acl_resource_type  NOT NULL,
    resource_id     UUID NOT NULL,
    principal_type  acl_principal_type NOT NULL,
    principal_id    UUID NOT NULL,
    permission      acl_permission     NOT NULL DEFAULT 'read',
    granted_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (resource_type, resource_id, principal_type, principal_id, permission)
);
CREATE INDEX idx_acl_grants_resource  ON acl_grants (org_id, resource_type, resource_id, permission);
CREATE INDEX idx_acl_grants_principal ON acl_grants (org_id, principal_type, principal_id);

ALTER TABLE documents ADD COLUMN is_restricted BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE folders   ADD COLUMN is_restricted BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_documents_restricted ON documents(id) WHERE is_restricted;
CREATE INDEX idx_folders_restricted   ON folders(id)   WHERE is_restricted;


-- Source: internal/migrate/sql/014_folder_index.sql
ALTER TABLE folders ADD COLUMN index_content TEXT;


