-- name: IsDocumentVisibleToUser :one
WITH RECURSIVE folder_chain(folder_id, parent_id, is_restricted) AS (
  SELECT f.id, f.parent_id, f.is_restricted
  FROM folders f JOIN documents d ON d.folder_id = f.id
  WHERE d.id = sqlc.arg(document_id)::uuid
    AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT pf.id, pf.parent_id, pf.is_restricted
  FROM folders pf JOIN folder_chain fc ON pf.id = fc.parent_id
)
SELECT
  sqlc.arg(is_admin)::boolean
  OR d.owner_id = sqlc.arg(user_id)::uuid
  OR (NOT d.is_restricted AND NOT EXISTS (SELECT 1 FROM folder_chain WHERE is_restricted))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission='read'
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='folder' AND g.resource_id IN (SELECT folder_id FROM folder_chain)
         AND g.permission='read'
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))) AS visible
FROM documents d
WHERE d.id = sqlc.arg(document_id)::uuid
  AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid;

-- name: ListGroupIDsForUser :many
SELECT gm.group_id FROM group_members gm JOIN groups g ON g.id = gm.group_id
WHERE gm.user_id = sqlc.arg(user_id)::uuid AND g.org_id = sqlc.arg(org_id)::uuid;

-- name: ListVisibleDocuments :many
WITH RECURSIVE folder_ancestors AS (
  SELECT f.id AS origin_folder_id, f.id AS ancestor_id, f.parent_id, f.is_restricted
  FROM folders f WHERE f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT fa.origin_folder_id, pf.id, pf.parent_id, pf.is_restricted
  FROM folders pf JOIN folder_ancestors fa ON pf.id = fa.parent_id
)
SELECT d.id, d.tenant_id, d.org_id, d.folder_id, d.owner_id, d.title,
       d.doc_type, d.status, d.language, d.created_at
FROM documents d
WHERE d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  AND (sqlc.narg(doc_type)::text IS NULL OR d.doc_type::text = sqlc.narg(doc_type)::text)
  AND (sqlc.narg(folder_id)::uuid IS NULL OR d.folder_id = sqlc.narg(folder_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR d.status::text = sqlc.narg(status)::text)
  AND (
    sqlc.arg(is_admin)::boolean
    OR d.owner_id = sqlc.arg(user_id)::uuid
    OR (NOT d.is_restricted AND NOT EXISTS (
         SELECT 1 FROM folder_ancestors fa WHERE fa.origin_folder_id=d.folder_id AND fa.is_restricted))
    OR EXISTS (SELECT 1 FROM acl_grants g
         WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission='read'
           AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
             OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
    OR EXISTS (SELECT 1 FROM folder_ancestors fa
         JOIN acl_grants g ON g.resource_type='folder' AND g.resource_id=fa.ancestor_id AND g.permission='read'
         WHERE fa.origin_folder_id=d.folder_id
           AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
             OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))))
  ORDER BY d.created_at DESC
  LIMIT sqlc.arg(limit_count);
