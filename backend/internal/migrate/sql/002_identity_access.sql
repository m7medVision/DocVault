-- +goose Up

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

-- +goose Down
DROP TABLE IF EXISTS memberships;
DROP TYPE IF EXISTS membership_role;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS tenants;
