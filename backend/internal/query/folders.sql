-- name: CreateFolder :exec
INSERT INTO folders (id, tenant_id, org_id, parent_id, name, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: GetFolderByID :one
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted
FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListFoldersByParent :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted
FROM folders
WHERE tenant_id = $1 AND org_id = $2 AND parent_id = $3
ORDER BY name ASC;

-- name: ListRootFolders :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted
FROM folders
WHERE tenant_id = $1 AND org_id = $2 AND parent_id IS NULL
ORDER BY name ASC;

-- name: ListAllFolders :many
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted
FROM folders
WHERE tenant_id = $1 AND org_id = $2
ORDER BY name ASC;

-- name: UpdateFolder :exec
UPDATE folders
SET name = $1, parent_id = $2
WHERE id = $3 AND tenant_id = $4 AND org_id = $5;

-- name: GetFolderByParentName :one
-- Finds a folder by its (tenant, org, parent, name) tuple. A NULL parent_id
-- matches root-level folders. Name matching is case-insensitive to mirror the
-- unique-name indexes (lower(name)).
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted
FROM folders
WHERE tenant_id = $1 AND org_id = $2
  AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)::uuid
  AND lower(name) = lower(sqlc.arg(name)::text)
LIMIT 1;

-- name: MoveFolder :execrows
-- Reparents a folder while preserving its name. A NULL parent_id moves the
-- folder to root.
UPDATE folders
SET parent_id = sqlc.narg(parent_id)::uuid
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;

-- name: GetFolderAncestorIDs :many
-- Returns the folder itself plus all of its ancestors (walking parent_id up to
-- the root), cycle-protected via a visited path. Used to detect reparent cycles:
-- a folder may not be moved under itself or any of its descendants, which is
-- equivalent to rejecting when the moved folder appears among the target
-- parent's ancestors (inclusive of the target parent).
WITH RECURSIVE ancestors(id, parent_id, path) AS (
  SELECT f.id, f.parent_id, ARRAY[f.id]
  FROM folders f
  WHERE f.id = sqlc.arg(folder_id)::uuid
    AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT pf.id, pf.parent_id, a.path || pf.id
  FROM folders pf JOIN ancestors a ON pf.id = a.parent_id
  WHERE NOT pf.id = ANY(a.path) AND array_length(a.path, 1) < 100
)
SELECT id FROM ancestors;

-- name: DeleteFolder :execrows
DELETE FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;
