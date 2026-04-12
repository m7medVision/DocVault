package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// auditRepository handles audit event data access.
type auditRepository struct {
	db *pgxpool.Pool
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return &auditRepository{db: db}
}

// Create creates a new audit event.
func (r *auditRepository) Create(ctx context.Context, event *model.AuditEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal audit metadata: %w", err)
	}

	query := `
		INSERT INTO audit_events (id, tenant_id, actor_id, entity_type, entity_id, action, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err = r.db.Exec(ctx, query,
		event.ID, event.TenantID, event.ActorID,
		event.EntityType, event.EntityID, event.Action, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}
	return nil
}

// ListByTenant lists audit events for a tenant with filters and cursor pagination.
func (r *auditRepository) ListByTenant(ctx context.Context, tenantID string, entityType, actorID, action, cursor string, limit int) ([]model.AuditEvent, *string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, tenant_id, actor_id, entity_type, entity_id, action, metadata, created_at
		FROM audit_events
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argCount := 1

	if entityType != "" {
		argCount++
		query += fmt.Sprintf(" AND entity_type = $%d", argCount)
		args = append(args, entityType)
	}
	if actorID != "" {
		argCount++
		query += fmt.Sprintf(" AND actor_id = $%d", argCount)
		args = append(args, actorID)
	}
	if action != "" {
		argCount++
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, action)
	}
	if cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM audit_events WHERE id = $%d)", argCount)
		args = append(args, cursor)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	events := make([]model.AuditEvent, 0)
	for rows.Next() {
		var e model.AuditEvent
		var metadataJSON []byte
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.ActorID, &e.EntityType,
			&e.EntityID, &e.Action, &metadataJSON, &e.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan audit event: %w", err)
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &e.Metadata); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal audit metadata: %w", err)
			}
		}
		events = append(events, e)
	}

	var resultCursor *string
	if len(events) > limit {
		events = events[:limit]
		c := events[len(events)-1].ID
		resultCursor = &c
	}

	return events, resultCursor, nil
}
