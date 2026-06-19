package repository

import (
	"context"
	"time"

	model "github.com/docvault/backend/internal/domain/document"
)

// VisibilityParams identifies a single document and the principal whose read
// visibility is being evaluated.
type VisibilityParams struct {
	TenantID   string
	OrgID      string
	DocumentID string
	UserID     string
	GroupIDs   []string
	IsAdmin    bool
}

// FolderVisibilityParams identifies a single folder and the principal whose read
// visibility is being evaluated.
type FolderVisibilityParams struct {
	TenantID string
	OrgID    string
	FolderID string
	UserID   string
	GroupIDs []string
	IsAdmin  bool
}

// ListVisibleParams filters a per-principal visible document listing.
type ListVisibleParams struct {
	TenantID string
	OrgID    string
	UserID   string
	GroupIDs []string
	IsAdmin  bool
	DocType  string
	FolderID string
	Status   string
	Language string
	Cursor   string
	Limit    int
}

// Group is an org-scoped principal grouping.
type Group struct {
	ID        string
	TenantID  string
	OrgID     string
	Name      string
	CreatedBy *string
	CreatedAt time.Time
}

// Grant is a per-resource read/write/delete permission for a principal.
type Grant struct {
	ID            string
	TenantID      string
	OrgID         string
	ResourceType  string
	ResourceID    string
	PrincipalType string
	PrincipalID   string
	Permission    string
	GrantedBy     *string
	CreatedAt     time.Time
}

// CreateGroupParams creates an org-scoped group.
type CreateGroupParams struct {
	TenantID  string
	OrgID     string
	Name      string
	CreatedBy *string
}

// CreateGrantParams creates a per-resource grant for a principal.
type CreateGrantParams struct {
	TenantID      string
	OrgID         string
	ResourceType  string
	ResourceID    string
	PrincipalType string
	PrincipalID   string
	Permission    string
	GrantedBy     *string
}

// ResourceRef identifies a single ACL-managed resource within an org.
type ResourceRef struct {
	TenantID     string
	OrgID        string
	ResourceType string
	ResourceID   string
}

// SetRestrictedParams toggles the is_restricted flag on a document or folder.
type SetRestrictedParams struct {
	TenantID   string
	OrgID      string
	ResourceID string
	Restricted bool
}

// The ACL surface is segregated into role-specific ports so consumers can
// depend on just the slice they use (interface segregation). The concrete
// adapter satisfies every port, and ACLRepository composes them for the
// composition root and the few callers that genuinely span roles.

// DocVisibilityPort evaluates per-document read/write visibility for a
// principal. The retrieval/authorization seam depends on just this.
type DocVisibilityPort interface {
	IsDocumentVisible(ctx context.Context, params VisibilityParams) (bool, error)
	IsDocumentWritable(ctx context.Context, params VisibilityParams) (bool, error)
}

// FolderVisibilityPort evaluates per-folder read visibility for a principal.
type FolderVisibilityPort interface {
	IsFolderVisible(ctx context.Context, params FolderVisibilityParams) (bool, error)
}

// GroupMembershipPort resolves the groups a user belongs to within an org. It
// is read on nearly every request that evaluates visibility.
type GroupMembershipPort interface {
	ListUserGroupIDs(ctx context.Context, userID, orgID string) ([]string, error)
}

// DocumentListingPort lists the documents a principal may read.
type DocumentListingPort interface {
	ListVisibleDocuments(ctx context.Context, params ListVisibleParams) ([]model.Document, *string, error)
}

// GroupAdminPort manages org-scoped groups and their membership.
type GroupAdminPort interface {
	CreateGroup(ctx context.Context, params CreateGroupParams) (Group, error)
	DeleteGroup(ctx context.Context, tenantID, orgID, groupID string) (int64, error)
	ListGroups(ctx context.Context, tenantID, orgID string) ([]Group, error)
	AddGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) error
	RemoveGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) (int64, error)
}

// GrantPort manages per-resource permission grants.
type GrantPort interface {
	CreateGrant(ctx context.Context, params CreateGrantParams) (string, error)
	GrantTargetExists(ctx context.Context, ref ResourceRef) (bool, error)
	DeleteGrant(ctx context.Context, tenantID, orgID, grantID string) (int64, error)
	ListGrantsByResource(ctx context.Context, ref ResourceRef) ([]Grant, error)
	DeleteGrantsForResource(ctx context.Context, ref ResourceRef) (int64, error)
}

// RestrictionPort toggles the is_restricted flag on documents and folders.
type RestrictionPort interface {
	SetDocumentRestricted(ctx context.Context, params SetRestrictedParams) (int64, error)
	SetFolderRestricted(ctx context.Context, params SetRestrictedParams) (int64, error)
}

// ACLRepository composes every ACL role port. It is the producer-side superset
// satisfied by the concrete postgres adapter; prefer depending on the narrow
// role ports above at each consumer.
type ACLRepository interface {
	DocVisibilityPort
	FolderVisibilityPort
	GroupMembershipPort
	DocumentListingPort
	GroupAdminPort
	GrantPort
	RestrictionPort
}
