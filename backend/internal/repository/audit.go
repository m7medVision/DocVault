package repository

import (
	"context"
	"encoding/json"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// auditRepository handles audit event data access.
type auditRepository struct {
	queries sqldb.Querier
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return &auditRepository{queries: sqldb.New(db)}
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

	err = r.queries.CreateAuditEvent(ctx, sqldb.CreateAuditEventParams{
		ID:         event.ID,
		TenantID:   event.TenantID,
		ActorID:    event.ActorID,
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		Action:     event.Action,
		Metadata:   metadataJSON,
	})
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

	var entityTypeFilter *string
	if entityType != "" {
		entityTypeFilter = &entityType
	}
	var actorIDFilter *string
	if actorID != "" {
		actorIDFilter = &actorID
	}
	var actionFilter *string
	if action != "" {
		actionFilter = &action
	}
	var cursorFilter *string
	if cursor != "" {
		cursorFilter = &cursor
	}

	rows, err := r.queries.ListAuditEvents(ctx, sqldb.ListAuditEventsParams{
		TenantIDArg: tenantID,
		EntityType:  entityTypeFilter,
		ActorID:     actorIDFilter,
		Action:      actionFilter,
		CursorID:    cursorFilter,
		LimitCount:  int32(limit + 1),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list audit events: %w", err)
	}

	events := make([]model.AuditEvent, 0, len(rows))
	for _, row := range rows {
		event, err := toModelAuditEvent(row)
		if err != nil {
			return nil, nil, err
		}
		events = append(events, event)
	}

	var resultCursor *string
	if len(events) > limit {
		events = events[:limit]
		c := events[len(events)-1].ID
		resultCursor = &c
	}

	return events, resultCursor, nil
}

func toModelAuditEvent(event sqldb.AuditEvent) (model.AuditEvent, error) {
	modelEvent := model.AuditEvent{
		ID:         event.ID,
		TenantID:   event.TenantID,
		ActorID:    event.ActorID,
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		Action:     event.Action,
		CreatedAt:  event.CreatedAt.Time,
	}
	if len(event.Metadata) > 0 {
		if err := json.Unmarshal(event.Metadata, &modelEvent.Metadata); err != nil {
			return model.AuditEvent{}, fmt.Errorf("failed to unmarshal audit metadata: %w", err)
		}
	}
	return modelEvent, nil
}
