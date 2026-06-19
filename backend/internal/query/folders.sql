-- name: CreateFolder :exec
INSERT INTO folders (id, tenant_id, org_id, parent_id, name, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: GetFolderByID :one
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted, index_content
FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListFoldersByParent :many
-- List endpoints intentionally omit index_content: the markdown "About this
-- folder" overview is large, never serialized in lists (json:"-"), and exposed
-- only via the dedicated, visibility-checked folder index endpoint.
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
SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at, is_restricted, index_content
FROM folders
WHERE tenant_id = $1 AND org_id = $2
  AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)::uuid
  AND lower(name) = lower(sqlc.arg(name)::text)
LIMIT 1;

-- name: MoveFolder :execrows
-- Reparents a folder while preserving its name. A NULL new_parent moves the
-- folder to root. The cycle check is performed atomically inside this single
-- UPDATE so two concurrent opposite reparents cannot each pass a separate check
-- and then both write a cycle: the update only succeeds when the new parent is
-- neither the folder itself nor any of its descendants. Descendants are gathered
-- by a recursive CTE seeded at the folder being moved, cycle-protected with a
-- path accumulator and a depth cap mirroring the other recursive queries. A
-- 0 rows-affected result signals a rejected (cycle/invalid-parent) move.
WITH RECURSIVE descendants(descendant_id, path) AS (
  SELECT f.id, ARRAY[f.id]
  FROM folders f
  WHERE f.id = sqlc.arg(id)::uuid
    AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT cf.id, d.path || cf.id
  FROM folders cf JOIN descendants d ON cf.parent_id = d.descendant_id
  WHERE cf.tenant_id = sqlc.arg(tenant_id)::uuid AND cf.org_id = sqlc.arg(org_id)::uuid
    AND NOT cf.id = ANY(d.path) AND array_length(d.path, 1) < 100
)
UPDATE folders
SET parent_id = sqlc.narg(new_parent)::uuid
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid
  AND (
    sqlc.narg(new_parent)::uuid IS NULL
    OR (
      sqlc.narg(new_parent)::uuid <> id
      AND sqlc.narg(new_parent)::uuid NOT IN (SELECT descendant_id FROM descendants)
    )
  );

-- name: GetFolderSubtreeHeight :one
-- Returns the height of the subtree rooted at folder_id, i.e. the number of
-- folder levels from the folder itself down to its deepest descendant. The
-- folder alone is height 1. Used as a soft depth guard so a reparent that would
-- push the moved subtree's deepest leaf past the cap is rejected. Cycle-protected
-- with a path accumulator and a depth cap mirroring the other recursive queries.
WITH RECURSIVE subtree(id, path, depth) AS (
  SELECT f.id, ARRAY[f.id], 1
  FROM folders f
  WHERE f.id = sqlc.arg(folder_id)::uuid
    AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT cf.id, s.path || cf.id, s.depth + 1
  FROM folders cf JOIN subtree s ON cf.parent_id = s.id
  WHERE cf.tenant_id = sqlc.arg(tenant_id)::uuid AND cf.org_id = sqlc.arg(org_id)::uuid
    AND NOT cf.id = ANY(s.path) AND array_length(s.path, 1) < 100
)
SELECT COALESCE(MAX(depth), 0)::int AS height FROM subtree;

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
  WHERE pf.tenant_id = sqlc.arg(tenant_id)::uuid AND pf.org_id = sqlc.arg(org_id)::uuid
    AND NOT pf.id = ANY(a.path) AND array_length(a.path, 1) < 100
)
SELECT id FROM ancestors;

-- name: DeleteFolder :execrows
DELETE FROM folders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: SetFolderIndex :execrows
UPDATE folders
SET index_content = sqlc.narg(index_content)::text
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;

-- name: GetFolderIndex :one
SELECT index_content
FROM folders
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;
