-- +goose Up
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

-- +goose Down
DROP INDEX IF EXISTS idx_folders_restricted;
DROP INDEX IF EXISTS idx_documents_restricted;
ALTER TABLE folders   DROP COLUMN IF EXISTS is_restricted;
ALTER TABLE documents DROP COLUMN IF EXISTS is_restricted;
DROP TABLE IF EXISTS acl_grants;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TYPE IF EXISTS acl_permission;
DROP TYPE IF EXISTS acl_principal_type;
DROP TYPE IF EXISTS acl_resource_type;
