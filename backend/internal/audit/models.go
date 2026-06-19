package audit

import "time"

// AuditEvent records an action in the system (append-only).
type AuditEvent struct {
	ID         string                 `json:"id"`
	TenantID   string                 `json:"tenant_id"`
	ActorID    *string                `json:"actor_id,omitempty"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
