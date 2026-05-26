-- name: CreateReminderRule :exec
INSERT INTO reminder_rules (id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW());

-- name: GetReminderRuleByID :one
SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
FROM reminder_rules
WHERE id = $1 AND tenant_id = $2;

-- name: GetReminderRulesByDocument :many
SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
FROM reminder_rules
WHERE document_id = $1 AND tenant_id = $2
ORDER BY trigger_date ASC;

-- name: ListReminderRulesByTenant :many
SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
FROM reminder_rules
WHERE tenant_id = $1
ORDER BY trigger_date ASC, created_at DESC;

-- name: ListActiveReminderRulesByTenant :many
SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
FROM reminder_rules
WHERE tenant_id = $1 AND active = true
ORDER BY trigger_date ASC, created_at DESC;

-- name: ListUpcomingReminderRules :many
SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
FROM reminder_rules
WHERE tenant_id = $1
  AND active = true
  AND trigger_date <= CURRENT_DATE + (sqlc.arg(within_days)::integer)
ORDER BY trigger_date ASC;

-- name: UpdateReminderRule :exec
UPDATE reminder_rules
SET rule_type = $1, trigger_date = $2, notify_days_before = $3, active = $4
WHERE id = $5 AND tenant_id = $6;

-- name: DeleteReminderRule :execrows
DELETE FROM reminder_rules
WHERE id = $1 AND tenant_id = $2;

-- name: CreateReminderEvent :exec
INSERT INTO reminder_events (id, rule_id, scheduled_at, sent_at, channel, status, error_message, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());

-- name: UpdateReminderEvent :exec
UPDATE reminder_events
SET sent_at = $1, status = $2, error_message = $3
WHERE id = $4;

-- name: GetPendingReminderEvents :many
SELECT id, rule_id, scheduled_at, sent_at, channel, status, error_message, created_at
FROM reminder_events
WHERE status = 'pending' AND scheduled_at <= $1
ORDER BY scheduled_at ASC;
