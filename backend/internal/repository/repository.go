// Package repository handles data access to PostgreSQL and MinIO.
// Repositories provide an abstraction over storage backends.
package repository

import (
	"context"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/domain/audit"
	"github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/domain/identity"
	"github.com/docvault/backend/internal/domain/notification"
	"github.com/docvault/backend/internal/domain/reminder"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentRepository provides document data access.
type DocumentRepository interface {
	Create(ctx context.Context, doc *document.Document) error
	GetByID(ctx context.Context, tenantID, orgID, id string) (*document.Document, error)
	List(ctx context.Context, q *ListDocumentsQuery) ([]document.Document, *string, error)
	Update(ctx context.Context, doc *document.Document) error
	Delete(ctx context.Context, tenantID, orgID, id, actorID string) error
	CreateVersion(ctx context.Context, version *document.DocumentVersion) error
	GetVersions(ctx context.Context, tenantID, documentID string) ([]document.DocumentVersion, error)
	CreatePage(ctx context.Context, page *document.DocumentPage) error
	GetPages(ctx context.Context, tenantID, documentID string) ([]document.DocumentPage, error)
	SetMetadata(ctx context.Context, tenantID string, metadata *document.DocumentMetadata) error
	GetMetadata(ctx context.Context, tenantID, documentID string) ([]document.DocumentMetadata, error)
	UpdateMetadataField(ctx context.Context, tenantID, documentID, key, correctedValue, correctedBy string) error
	GetFullDocument(ctx context.Context, tenantID, orgID, documentID string) (*document.Document, []document.DocumentVersion, []document.DocumentMetadata, error)
	UpdateProcessingFields(ctx context.Context, tenantID, documentID string, stage *string, errMsg *string) error
	ClearSuggestion(ctx context.Context, tenantID, orgID, documentID string, stage *string) error
	GetStats(ctx context.Context, tenantID, orgID string) (*document.DocumentStats, error)
}

// ReminderRepository provides reminder data access.
type ReminderRepository interface {
	Create(ctx context.Context, reminder *reminder.ReminderRule) error
	GetByID(ctx context.Context, tenantID, id string) (*reminder.ReminderRule, error)
	GetByDocument(ctx context.Context, tenantID, documentID string) ([]reminder.ReminderRule, error)
	ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]reminder.ReminderRule, error)
	ListUpcoming(ctx context.Context, tenantID string, withinDays int) ([]reminder.ReminderRule, error)
	Update(ctx context.Context, reminder *reminder.ReminderRule) error
	Delete(ctx context.Context, tenantID, id string) error
	CreateEvent(ctx context.Context, event *reminder.ReminderEvent) error
	UpdateEvent(ctx context.Context, event *reminder.ReminderEvent) error
	GetPendingEvents(ctx context.Context, before time.Time) ([]reminder.ReminderEvent, error)
}

// FolderRepository provides folder data access.
type FolderRepository interface {
	Create(ctx context.Context, folder *document.Folder) error
	GetByID(ctx context.Context, tenantID, orgID, id string) (*document.Folder, error)
	GetByParentName(ctx context.Context, tenantID, orgID string, parentID *string, name string) (*document.Folder, error)
	ListByParent(ctx context.Context, tenantID, orgID, parentID string) ([]document.Folder, error)
	ListRoot(ctx context.Context, tenantID, orgID string) ([]document.Folder, error)
	ListAll(ctx context.Context, tenantID, orgID string) ([]document.Folder, error)
	GetAncestorIDs(ctx context.Context, tenantID, orgID, folderID string) ([]string, error)
	Update(ctx context.Context, folder *document.Folder) error
	Move(ctx context.Context, tenantID, orgID, folderID string, parentID *string) (int64, error)
	Delete(ctx context.Context, tenantID, orgID, id string) error
}

// ErrFolderNameExists is returned by folder Create-style operations when a
// folder with the same (tenant, org, parent, name) already exists. Callers that
// find-or-create (e.g. EnsureFolderPath) treat it as a signal to fetch the
// existing folder rather than fail.
var ErrFolderNameExists = errFolderNameExists

// TagRepository provides tag data access.
type TagRepository interface {
	Create(ctx context.Context, tag *document.Tag) error
	GetByID(ctx context.Context, tenantID, id string) (*document.Tag, error)
	GetByName(ctx context.Context, tenantID, name string) (*document.Tag, error)
	List(ctx context.Context, tenantID string, query string, limit int) ([]document.Tag, error)
	Delete(ctx context.Context, tenantID, id string) error
	AddToDocument(ctx context.Context, tenantID, tagID, documentID string) error
	RemoveFromDocument(ctx context.Context, tenantID, tagID, documentID string) error
	GetDocumentTags(ctx context.Context, tenantID, documentID string) ([]document.Tag, error)
}

// AuditRepository provides audit data access.
type AuditRepository interface {
	Create(ctx context.Context, event *audit.AuditEvent) error
	ListByTenant(ctx context.Context, tenantID string, entityType, actorID, action, cursor string, limit int) ([]audit.AuditEvent, *string, error)
}

// NotificationRepository provides notification data access.
type NotificationRepository interface {
	Create(ctx context.Context, notification *notification.Notification) error
	List(ctx context.Context, tenantID, userID string, status notification.NotificationStatus, cursor string, limit int) ([]notification.Notification, *string, error)
	MarkRead(ctx context.Context, tenantID, userID, notificationID string) error
	GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error)
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*identity.User, error)
	FindByID(ctx context.Context, id string) (*identity.User, error)
	Create(ctx context.Context, u *identity.User) error
	UpdateProfile(ctx context.Context, userID, displayName, locale string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error
	IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error)
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
	ACL          ACLRepository
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
		ACL:          NewACLRepository(db),
	}
}
