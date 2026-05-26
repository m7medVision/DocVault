-- name: CreateTenant :exec
INSERT INTO tenants (id, name, plan, created_at)
VALUES ($1, $2, $3, $4);

-- name: CreateOrganization :exec
INSERT INTO organizations (id, tenant_id, name, created_at)
VALUES ($1, $2, $3, $4);
