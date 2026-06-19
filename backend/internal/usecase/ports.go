package usecase

import (
	"context"

	model "github.com/docvault/backend/internal/document"
	remindermodel "github.com/docvault/backend/internal/reminder"
	"github.com/docvault/backend/internal/repository"
)

// This file collects the narrow, consumer-side persistence ports the
// application services depend on. Following interface segregation, each service
// depends only on the methods it actually calls: the broad repository.*
// interfaces are the producer-side superset, and the concrete repositories
// satisfy these narrow views automatically. New services should declare their
// own port here rather than reach for a fat repository interface.

// GrantCleaner removes every ACL grant attached to a resource. Document and
// folder deletion use it to avoid leaving orphaned grants behind once the
// resource is gone. It is the single ACL method FolderService needs.
type GrantCleaner interface {
	DeleteGrantsForResource(ctx context.Context, ref repository.ResourceRef) (int64, error)
}

// DocumentStore is the persistence port DocumentService depends on: the subset
// of repository.DocumentRepository it actually calls. Listing is intentionally
// absent — DocumentService.List goes through the ACL-aware DocumentACL port so
// the visibility predicate is never bypassed.
type DocumentStore interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Document, error)
	Update(ctx context.Context, doc *model.Document) error
	Delete(ctx context.Context, tenantID, orgID, id, actorID string) error
	CreateVersion(ctx context.Context, version *model.DocumentVersion) error
	GetVersions(ctx context.Context, tenantID, documentID string) ([]model.DocumentVersion, error)
	GetPages(ctx context.Context, tenantID, documentID string) ([]model.DocumentPage, error)
	GetFullDocument(ctx context.Context, tenantID, orgID, documentID string) (*model.Document, []model.DocumentVersion, []model.DocumentMetadata, error)
	UpdateMetadataField(ctx context.Context, tenantID, documentID, key, correctedValue, correctedBy string) error
	UpdateProcessingFields(ctx context.Context, tenantID, documentID string, stage *string, errMsg *string) error
	ClearSuggestion(ctx context.Context, tenantID, orgID, documentID string, stage *string) error
	ApplySuggestion(ctx context.Context, doc *model.Document, stage *string) error
	GetStats(ctx context.Context, tenantID, orgID string) (*model.DocumentStats, error)
}

// DocumentVisibilityLister lists the documents a principal may read, applying
// the per-row visibility predicate. It is the ACL view DocumentService.List
// depends on.
type DocumentVisibilityLister interface {
	ListVisibleDocuments(ctx context.Context, params repository.ListVisibleParams) ([]model.Document, *string, error)
}

// DocumentACL is the combined ACL view DocumentService needs: list the
// documents a principal may read, and clean up a document's grants on delete.
type DocumentACL interface {
	DocumentVisibilityLister
	GrantCleaner
}

// ReminderStore is the persistence port the HTTP-facing ReminderService needs.
// It is the rule-management subset of repository.ReminderRepository; the event
// dispatch methods (CreateEvent/UpdateEvent/GetPendingEvents/ListUpcoming) are
// driven by the separate reminder worker, not this service.
type ReminderStore interface {
	Create(ctx context.Context, rule *remindermodel.ReminderRule) error
	GetByID(ctx context.Context, tenantID, id string) (*remindermodel.ReminderRule, error)
	GetByDocument(ctx context.Context, tenantID, documentID string) ([]remindermodel.ReminderRule, error)
	ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]remindermodel.ReminderRule, error)
	Update(ctx context.Context, rule *remindermodel.ReminderRule) error
}
