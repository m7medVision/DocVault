-- name: CreateTag :exec
INSERT INTO tags (id, tenant_id, name, created_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (tenant_id, name) DO NOTHING;

-- name: GetTagByID :one
SELECT id, tenant_id, name, created_at
FROM tags
WHERE id = $1 AND tenant_id = $2;

-- name: GetTagByName :one
SELECT id, tenant_id, name, created_at
FROM tags
WHERE tenant_id = $1 AND name = $2;

-- name: ListTags :many
SELECT id, tenant_id, name, created_at
FROM tags
WHERE tenant_id = $1
ORDER BY name ASC
LIMIT $2;

-- name: SearchTags :many
SELECT id, tenant_id, name, created_at
FROM tags
WHERE tenant_id = $1 AND name ILIKE $2
ORDER BY name ASC
LIMIT $3;

-- name: DeleteTag :execrows
DELETE FROM tags
WHERE id = $1 AND tenant_id = $2;

-- name: AddTagToDocument :exec
INSERT INTO document_tags (document_id, tag_id, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromDocument :exec
DELETE FROM document_tags
WHERE document_id = $1 AND tag_id = $2;

-- name: GetDocumentTags :many
SELECT t.id, t.tenant_id, t.name, t.created_at
FROM tags t
JOIN document_tags dt ON t.id = dt.tag_id
WHERE dt.document_id = $1 AND t.tenant_id = $2
ORDER BY t.name ASC;
