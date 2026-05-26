-- name: CreateAuditEvent :exec
INSERT INTO audit_events (id, tenant_id, actor_id, entity_type, entity_id, action, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());

-- name: ListAuditEvents :many
SELECT ae.id, ae.tenant_id, ae.actor_id, ae.entity_type, ae.entity_id, ae.action, ae.metadata, ae.created_at
FROM audit_events ae
WHERE ae.tenant_id = sqlc.arg(tenant_id_arg)
  AND (sqlc.narg(entity_type)::text IS NULL OR ae.entity_type = sqlc.narg(entity_type)::text)
  AND (sqlc.narg(actor_id)::uuid IS NULL OR ae.actor_id = sqlc.narg(actor_id)::uuid)
  AND (sqlc.narg(action)::text IS NULL OR ae.action = sqlc.narg(action)::text)
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR ae.created_at < (SELECT created_at FROM audit_events WHERE id = sqlc.narg(cursor_id)::uuid))
ORDER BY ae.created_at DESC
LIMIT sqlc.arg(limit_count);
