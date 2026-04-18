// Package repository handles data access to PostgreSQL and MinIO.
// Repositories provide an abstraction over storage backends.
package repository

import (
	"context"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentRepository provides document data access.
type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Document, error)
	List(ctx context.Context, q *ListDocumentsQuery) ([]model.Document, *string, error)
	Update(ctx context.Context, doc *model.Document) error
	Delete(ctx context.Context, tenantID, orgID, id, actorID string) error
	CreateVersion(ctx context.Context, version *model.DocumentVersion) error
	GetVersions(ctx context.Context, tenantID, documentID string) ([]model.DocumentVersion, error)
	CreatePage(ctx context.Context, page *model.DocumentPage) error
	GetPages(ctx context.Context, tenantID, documentID string) ([]model.DocumentPage, error)
	SetMetadata(ctx context.Context, tenantID string, metadata *model.DocumentMetadata) error
	GetMetadata(ctx context.Context, tenantID, documentID string) ([]model.DocumentMetadata, error)
	UpdateMetadataField(ctx context.Context, tenantID, documentID, key, correctedValue, correctedBy string) error
	GetFullDocument(ctx context.Context, tenantID, orgID, documentID string) (*model.Document, []model.DocumentVersion, []model.DocumentMetadata, error)
	UpdateProcessingFields(ctx context.Context, tenantID, documentID string, stage *string, errMsg *string) error
}

// ReminderRepository provides reminder data access.
type ReminderRepository interface {
	Create(ctx context.Context, reminder *model.ReminderRule) error
	GetByID(ctx context.Context, tenantID, id string) (*model.ReminderRule, error)
	GetByDocument(ctx context.Context, tenantID, documentID string) ([]model.ReminderRule, error)
	ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]model.ReminderRule, error)
	ListUpcoming(ctx context.Context, tenantID string, withinDays int) ([]model.ReminderRule, error)
	Update(ctx context.Context, reminder *model.ReminderRule) error
	Delete(ctx context.Context, tenantID, id string) error
	CreateEvent(ctx context.Context, event *model.ReminderEvent) error
	UpdateEvent(ctx context.Context, event *model.ReminderEvent) error
	GetPendingEvents(ctx context.Context, before time.Time) ([]model.ReminderEvent, error)
}

// FolderRepository provides folder data access.
type FolderRepository interface {
	Create(ctx context.Context, folder *model.Folder) error
	GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Folder, error)
	ListByParent(ctx context.Context, tenantID, orgID, parentID string) ([]model.Folder, error)
	ListRoot(ctx context.Context, tenantID, orgID string) ([]model.Folder, error)
	ListAll(ctx context.Context, tenantID, orgID string) ([]model.Folder, error)
	Update(ctx context.Context, folder *model.Folder) error
	Delete(ctx context.Context, tenantID, orgID, id string) error
}

// TagRepository provides tag data access.
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetByID(ctx context.Context, tenantID, id string) (*model.Tag, error)
	GetByName(ctx context.Context, tenantID, name string) (*model.Tag, error)
	List(ctx context.Context, tenantID string, query string, limit int) ([]model.Tag, error)
	Delete(ctx context.Context, tenantID, id string) error
	AddToDocument(ctx context.Context, tenantID, tagID, documentID string) error
	RemoveFromDocument(ctx context.Context, tenantID, tagID, documentID string) error
	GetDocumentTags(ctx context.Context, tenantID, documentID string) ([]model.Tag, error)
}

// AuditRepository provides audit data access.
type AuditRepository interface {
	Create(ctx context.Context, event *model.AuditEvent) error
	ListByTenant(ctx context.Context, tenantID string, entityType, actorID, action, cursor string, limit int) ([]model.AuditEvent, *string, error)
}

// NotificationRepository provides notification data access.
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	List(ctx context.Context, tenantID, userID string, status model.NotificationStatus, cursor string, limit int) ([]model.Notification, *string, error)
	MarkRead(ctx context.Context, tenantID, userID, notificationID string) error
	GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error)
}

// Repositories holds all repository instances.
type Repositories struct {
	Document     DocumentRepository
	Reminder     ReminderRepository
	Folder       FolderRepository
	Tag          TagRepository
	Audit        AuditRepository
	Notification NotificationRepository
	User         UserRepository
	Membership   MembershipRepository
	Policy       PolicyRepository
	Search       SearchRepository
}

// NewRepositories creates all repository instances with the given database pool.
func NewRepositories(db *pgxpool.Pool, enforcer *casbin.Enforcer) *Repositories {
	return &Repositories{
		Document:     NewDocumentRepository(db),
		Reminder:     NewReminderRepository(db),
		Folder:       NewFolderRepository(db),
		Tag:          NewTagRepository(db),
		Audit:        NewAuditRepository(db),
		Notification: NewNotificationRepository(db),
		User:         NewUserRepository(db),
		Membership:   NewMembershipRepository(db),
		Policy:       NewPolicyRepository(enforcer),
		Search:       NewSearchRepository(db),
	}
}
