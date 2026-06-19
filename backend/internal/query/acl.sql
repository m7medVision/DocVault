-- name: IsDocumentVisibleToUser :one
WITH RECURSIVE folder_chain(folder_id, parent_id, is_restricted, path) AS (
  SELECT f.id, f.parent_id, f.is_restricted, ARRAY[f.id]
  FROM folders f JOIN documents d ON d.folder_id = f.id
  WHERE d.id = sqlc.arg(document_id)::uuid
    AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT pf.id, pf.parent_id, pf.is_restricted, fc.path || pf.id
  FROM folders pf JOIN folder_chain fc ON pf.id = fc.parent_id
  WHERE NOT pf.id = ANY(fc.path) AND array_length(fc.path, 1) < 100
)
SELECT
  sqlc.arg(is_admin)::boolean
  OR d.owner_id = sqlc.arg(user_id)::uuid
  OR (NOT d.is_restricted AND NOT EXISTS (SELECT 1 FROM folder_chain WHERE is_restricted))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission='read'
         AND g.org_id = sqlc.arg(org_id)::uuid
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='folder' AND g.resource_id IN (SELECT folder_id FROM folder_chain)
         AND g.permission='read'
         AND g.org_id = sqlc.arg(org_id)::uuid
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))) AS visible
FROM documents d
WHERE d.id = sqlc.arg(document_id)::uuid
  AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid;

-- name: IsDocumentWritableToUser :one
WITH RECURSIVE folder_chain(folder_id, parent_id, is_restricted, path) AS (
  SELECT f.id, f.parent_id, f.is_restricted, ARRAY[f.id]
  FROM folders f JOIN documents d ON d.folder_id = f.id
  WHERE d.id = sqlc.arg(document_id)::uuid
    AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT pf.id, pf.parent_id, pf.is_restricted, fc.path || pf.id
  FROM folders pf JOIN folder_chain fc ON pf.id = fc.parent_id
  WHERE NOT pf.id = ANY(fc.path) AND array_length(fc.path, 1) < 100
)
SELECT
  sqlc.arg(is_admin)::boolean
  OR d.owner_id = sqlc.arg(user_id)::uuid
  OR (NOT d.is_restricted AND NOT EXISTS (SELECT 1 FROM folder_chain WHERE is_restricted))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission IN ('write','delete')
         AND g.org_id = sqlc.arg(org_id)::uuid
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='folder' AND g.resource_id IN (SELECT folder_id FROM folder_chain)
         AND g.permission IN ('write','delete')
         AND g.org_id = sqlc.arg(org_id)::uuid
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))) AS writable
FROM documents d
WHERE d.id = sqlc.arg(document_id)::uuid
  AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid;

-- name: IsFolderVisibleToUser :one
-- Mirrors IsDocumentVisibleToUser but is seeded from the FOLDER itself instead of
-- a document's containing folder. The recursive folder_chain starts at the target
-- folder and walks parent_id up to the root, cycle-protected with a path
-- accumulator and a depth cap. The folder is visible iff the principal is an
-- admin, OR created the folder, OR the folder itself is NOT restricted and has no
-- restricted ancestor, OR there is a read grant on the folder or any of its
-- ancestors for the user or one of their groups (scoped to org_id).
WITH RECURSIVE folder_chain(folder_id, parent_id, is_restricted, path) AS (
  SELECT f.id, f.parent_id, f.is_restricted, ARRAY[f.id]
  FROM folders f
  WHERE f.id = sqlc.arg(folder_id)::uuid
    AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT pf.id, pf.parent_id, pf.is_restricted, fc.path || pf.id
  FROM folders pf JOIN folder_chain fc ON pf.id = fc.parent_id
  WHERE NOT pf.id = ANY(fc.path) AND array_length(fc.path, 1) < 100
)
SELECT
  sqlc.arg(is_admin)::boolean
  OR f.created_by = sqlc.arg(user_id)::uuid
  OR NOT EXISTS (SELECT 1 FROM folder_chain WHERE is_restricted)
  OR EXISTS (SELECT 1 FROM acl_grants g
       WHERE g.resource_type='folder' AND g.resource_id IN (SELECT folder_id FROM folder_chain)
         AND g.permission='read'
         AND g.org_id = sqlc.arg(org_id)::uuid
         AND ((g.principal_type='user'  AND g.principal_id=sqlc.arg(user_id)::uuid)
           OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))) AS visible
FROM folders f
WHERE f.id = sqlc.arg(folder_id)::uuid
  AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid;

-- name: ListGroupIDsForUser :many
SELECT gm.group_id FROM group_members gm JOIN groups g ON g.id = gm.group_id
WHERE gm.user_id = sqlc.arg(user_id)::uuid AND g.org_id = sqlc.arg(org_id)::uuid;

-- name: ListVisibleDocuments :many
WITH RECURSIVE candidate_folders AS (
  -- Seed the ancestor walk only from folders that hold a candidate document
  -- (same tenant/org and folder scoping the listing applies below), instead of
  -- every folder in the org, so the recursive cost stops scaling with the org's
  -- total folder count. This is a superset of the folders any listed document
  -- can resolve through, so the visibility decision below is unchanged.
  SELECT DISTINCT d.folder_id AS id
  FROM documents d
  WHERE d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
    AND d.folder_id IS NOT NULL
    AND (sqlc.narg(folder_id)::uuid IS NULL OR d.folder_id = sqlc.narg(folder_id)::uuid)
),
folder_ancestors AS (
  SELECT cf.id AS origin_folder_id, f.id AS ancestor_id, f.parent_id, f.is_restricted, ARRAY[f.id] AS path
  FROM candidate_folders cf JOIN folders f ON f.id = cf.id AND f.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT fa.origin_folder_id, pf.id, pf.parent_id, pf.is_restricted, fa.path || pf.id
  FROM folders pf JOIN folder_ancestors fa ON pf.id = fa.parent_id
  WHERE NOT pf.id = ANY(fa.path) AND array_length(fa.path, 1) < 100
)
SELECT d.id, d.tenant_id, d.org_id, d.folder_id, d.owner_id, d.title,
       d.doc_type, d.status, d.language, d.created_at
FROM documents d
WHERE d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  AND (sqlc.narg(doc_type)::text IS NULL OR d.doc_type::text = sqlc.narg(doc_type)::text)
  AND (sqlc.narg(folder_id)::uuid IS NULL OR d.folder_id = sqlc.narg(folder_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR d.status::text = sqlc.narg(status)::text)
  AND (sqlc.narg(language)::text IS NULL OR d.language = sqlc.narg(language)::text)
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR d.created_at < (SELECT created_at FROM documents WHERE id = sqlc.narg(cursor_id)::uuid))
  AND (
    sqlc.arg(is_admin)::boolean
    OR d.owner_id = sqlc.arg(user_id)::uuid
    OR (NOT d.is_restricted AND NOT EXISTS (
         SELECT 1 FROM folder_ancestors fa WHERE fa.origin_folder_id=d.folder_id AND fa.is_restricted))
    OR EXISTS (SELECT 1 FROM acl_grants g
         WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission='read'
           AND g.org_id = sqlc.arg(org_id)::uuid
           AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
             OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
    OR EXISTS (SELECT 1 FROM folder_ancestors fa
         JOIN acl_grants g ON g.resource_type='folder' AND g.resource_id=fa.ancestor_id AND g.permission='read'
           AND g.org_id = sqlc.arg(org_id)::uuid
         WHERE fa.origin_folder_id=d.folder_id
           AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
             OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))))
  ORDER BY d.created_at DESC
  LIMIT sqlc.arg(limit_count);

-- name: CreateGroup :one
INSERT INTO groups (tenant_id, org_id, name, created_by)
VALUES (sqlc.arg(tenant_id)::uuid, sqlc.arg(org_id)::uuid, sqlc.arg(name), sqlc.narg(created_by)::uuid)
RETURNING id, tenant_id, org_id, name, created_by, created_at;

-- name: DeleteGroup :execrows
DELETE FROM groups
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;

-- name: ListGroups :many
SELECT id, tenant_id, org_id, name, created_by, created_at
FROM groups
WHERE tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid
ORDER BY name ASC;

-- name: AddGroupMember :exec
INSERT INTO group_members (group_id, user_id)
SELECT sqlc.arg(group_id)::uuid, sqlc.arg(user_id)::uuid
FROM groups g
WHERE g.id = sqlc.arg(group_id)::uuid
  AND g.tenant_id = sqlc.arg(tenant_id)::uuid AND g.org_id = sqlc.arg(org_id)::uuid
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: RemoveGroupMember :execrows
DELETE FROM group_members gm
USING groups g
WHERE gm.group_id = g.id
  AND gm.group_id = sqlc.arg(group_id)::uuid AND gm.user_id = sqlc.arg(user_id)::uuid
  AND g.tenant_id = sqlc.arg(tenant_id)::uuid AND g.org_id = sqlc.arg(org_id)::uuid;

-- name: CreateGrant :one
INSERT INTO acl_grants (tenant_id, org_id, resource_type, resource_id, principal_type, principal_id, permission, granted_by)
VALUES (
  sqlc.arg(tenant_id)::uuid, sqlc.arg(org_id)::uuid,
  sqlc.arg(resource_type)::acl_resource_type, sqlc.arg(resource_id)::uuid,
  sqlc.arg(principal_type)::acl_principal_type, sqlc.arg(principal_id)::uuid,
  sqlc.arg(permission)::acl_permission, sqlc.narg(granted_by)::uuid)
ON CONFLICT (resource_type, resource_id, principal_type, principal_id, permission) DO NOTHING
RETURNING id;

-- name: GrantTargetExists :one
-- Returns true when the (resource_type, resource_id) refers to an existing
-- document or folder in the caller's tenant/org, used to validate a grant target.
SELECT EXISTS (
  SELECT 1 FROM documents d
  WHERE sqlc.arg(resource_type)::acl_resource_type = 'document'
    AND d.id = sqlc.arg(resource_id)::uuid
    AND d.tenant_id = sqlc.arg(tenant_id)::uuid AND d.org_id = sqlc.arg(org_id)::uuid
  UNION ALL
  SELECT 1 FROM folders f
  WHERE sqlc.arg(resource_type)::acl_resource_type = 'folder'
    AND f.id = sqlc.arg(resource_id)::uuid
    AND f.tenant_id = sqlc.arg(tenant_id)::uuid AND f.org_id = sqlc.arg(org_id)::uuid
) AS exists;

-- name: DeleteGrant :execrows
DELETE FROM acl_grants
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;

-- name: ListGrantsByResource :many
SELECT id, tenant_id, org_id, resource_type, resource_id, principal_type, principal_id, permission, granted_by, created_at
FROM acl_grants
WHERE tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid
  AND resource_type = sqlc.arg(resource_type)::acl_resource_type
  AND resource_id = sqlc.arg(resource_id)::uuid
ORDER BY created_at ASC;

-- name: DeleteGrantsForResource :execrows
DELETE FROM acl_grants
WHERE tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid
  AND resource_type = sqlc.arg(resource_type)::acl_resource_type
  AND resource_id = sqlc.arg(resource_id)::uuid;

-- name: SetDocumentRestricted :execrows
UPDATE documents
SET is_restricted = sqlc.arg(is_restricted)::boolean
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;

-- name: SetFolderRestricted :execrows
UPDATE folders
SET is_restricted = sqlc.arg(is_restricted)::boolean
WHERE id = sqlc.arg(id)::uuid
  AND tenant_id = sqlc.arg(tenant_id)::uuid AND org_id = sqlc.arg(org_id)::uuid;
