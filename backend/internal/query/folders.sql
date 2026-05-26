-- name: CreateFolder :exec
INSERT INTO folders (id, tenant_id, org_id, parent_id, name, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: GetFolderByID :one
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListFoldersByParent :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
FROM folders
WHERE tenant_id = $1 AND org_id = $2 AND parent_id = $3
ORDER BY name ASC;

-- name: ListRootFolders :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
FROM folders
WHERE tenant_id = $1 AND org_id = $2 AND parent_id IS NULL
ORDER BY name ASC;

-- name: ListAllFolders :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
FROM folders
WHERE tenant_id = $1 AND org_id = $2
ORDER BY name ASC;

-- name: UpdateFolder :exec
UPDATE folders
SET name = $1, parent_id = $2
WHERE id = $3 AND tenant_id = $4 AND org_id = $5;

-- name: DeleteFolder :execrows
DELETE FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;
