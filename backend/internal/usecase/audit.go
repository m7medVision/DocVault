// Package usecase contains application business logic.
package usecase

import (
	"context"
	"errors"
	"time"

	model "github.com/docvault/backend/internal/domain/audit"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrAuditTenantIDRequired   = errors.New("tenant_id is required")
	ErrAuditEntityTypeRequired = errors.New("entity_type is required")
	ErrAuditEntityIDRequired   = errors.New("entity_id is required")
	ErrAuditActionRequired     = errors.New("action is required")
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

type AuditAction string

const (
	AuditActionCreate   AuditAction = "create"
	AuditActionUpdate   AuditAction = "update"
	AuditActionDelete   AuditAction = "delete"
	AuditActionView     AuditAction = "view"
	AuditActionDownload AuditAction = "download"
	AuditActionShare    AuditAction = "share"
	AuditActionLogin    AuditAction = "login"
	AuditActionLogout   AuditAction = "logout"
)

type AuditLogger interface {
	Log(ctx context.Context, action AuditAction, tenantID, actorID, entityType, entityID string, metadata map[string]interface{})
}

// WriteAuditEventInput contains the data needed to write an audit event.
type WriteAuditEventInput struct {
	TenantID   string
	ActorID    *string
	EntityType string
	EntityID   string
	Action     AuditAction
	Metadata   map[string]interface{}
}

func (s *AuditService) Write(ctx context.Context, input *WriteAuditEventInput) error {
	if input.TenantID == "" {
		return ErrAuditTenantIDRequired
	}
	if input.EntityType == "" {
		return ErrAuditEntityTypeRequired
	}
	if input.EntityID == "" {
		return ErrAuditEntityIDRequired
	}
	if input.Action == "" {
		return ErrAuditActionRequired
	}

	event := model.AuditEvent{
		ID:         uuid.New().String(),
		TenantID:   input.TenantID,
		ActorID:    input.ActorID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Action:     string(input.Action),
		Metadata:   input.Metadata,
		CreatedAt:  time.Now(),
	}

	if s.repo == nil {
		return nil
	}

	return s.repo.Create(ctx, &event)
}

func (s *AuditService) Log(ctx context.Context, action AuditAction, tenantID, actorID, entityType, entityID string, metadata map[string]interface{}) {
	actorPtr := &actorID
	if actorID == "" {
		actorPtr = nil
	}
	s.Write(ctx, &WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    actorPtr,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Metadata:   metadata,
	})
}

// ListAuditEventsInput contains filters for listing audit events.
type ListAuditEventsInput struct {
	TenantID   string
	EntityType string
	EntityID   string
	ActorID    string
	Action     string
	StartDate  *time.Time
	EndDate    *time.Time
	Cursor     string
	Limit      int
}

// ListAuditEventsOutput contains the result of listing audit events.
type ListAuditEventsOutput struct {
	Events []model.AuditEvent
	Cursor *string
}

func (s *AuditService) List(ctx context.Context, input *ListAuditEventsInput) (*ListAuditEventsOutput, error) {
	if input.TenantID == "" {
		return nil, ErrAuditTenantIDRequired
	}

	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 50
	}

	if s.repo == nil {
		return &ListAuditEventsOutput{
			Events: []model.AuditEvent{},
			Cursor: nil,
		}, nil
	}

	events, cursor, err := s.repo.ListByTenant(ctx, input.TenantID, input.EntityType, input.ActorID, input.Action, input.Cursor, input.Limit)
	if err != nil {
		return nil, err
	}

	return &ListAuditEventsOutput{
		Events: events,
		Cursor: cursor,
	}, nil
}

func (s *AuditService) DocumentCreate(ctx context.Context, tenantID, actorID, documentID, docType string) {
	s.Log(ctx, AuditActionCreate, tenantID, actorID, "document", documentID, map[string]interface{}{
		"doc_type": docType,
	})
}

func (s *AuditService) DocumentDelete(ctx context.Context, tenantID, actorID, documentID string) {
	s.Log(ctx, AuditActionDelete, tenantID, actorID, "document", documentID, nil)
}

func (s *AuditService) DocumentDownload(ctx context.Context, tenantID, actorID, documentID string) {
	s.Log(ctx, AuditActionDownload, tenantID, actorID, "document", documentID, nil)
}

func (s *AuditService) MetadataUpdate(ctx context.Context, tenantID, actorID, documentID string, updatedFields []string) {
	s.Log(ctx, AuditActionUpdate, tenantID, actorID, "document", documentID, map[string]interface{}{
		"fields": updatedFields,
	})
}

func (s *AuditService) FolderCreate(ctx context.Context, tenantID, actorID, folderID string) {
	s.Log(ctx, AuditActionCreate, tenantID, actorID, "folder", folderID, nil)
}

func (s *AuditService) MemberInvite(ctx context.Context, tenantID, actorID, memberID, role string) {
	s.Log(ctx, AuditActionCreate, tenantID, actorID, "membership", memberID, map[string]interface{}{
		"role": role,
	})
}

func (s *AuditService) MemberRoleChange(ctx context.Context, tenantID, actorID, memberID, oldRole, newRole string) {
	s.Log(ctx, AuditActionUpdate, tenantID, actorID, "membership", memberID, map[string]interface{}{
		"old_role": oldRole,
		"new_role": newRole,
	})
}
