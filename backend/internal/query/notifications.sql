-- name: CreateNotification :exec
INSERT INTO notifications (id, tenant_id, user_id, type, title, body, link, status, metadata, created_at, read_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), $10);

-- name: ListNotifications :many
SELECT n.id, n.tenant_id, n.user_id, n.type, n.title, n.body, n.link, n.status, n.metadata, n.created_at, n.read_at
FROM notifications n
WHERE n.tenant_id = sqlc.arg(tenant_id_arg)
  AND n.user_id = sqlc.arg(user_id_arg)
  AND (sqlc.narg(status)::text IS NULL OR n.status = sqlc.narg(status)::text)
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR n.created_at < (SELECT created_at FROM notifications WHERE id = sqlc.narg(cursor_id)::uuid))
ORDER BY n.created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: MarkNotificationRead :execrows
UPDATE notifications
SET status = 'read', read_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND user_id = $3;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*)
FROM notifications
WHERE tenant_id = $1 AND user_id = $2 AND status = 'unread';
