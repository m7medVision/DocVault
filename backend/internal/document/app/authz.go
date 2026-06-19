package app

import (
	"context"

	"github.com/docvault/backend/internal/platform/apperr"
	"github.com/docvault/backend/internal/repository"
)

// Principal is the authenticated caller's security context for a single
// request, resolved once from the JWT claims at the transport edge.
type Principal struct {
	TenantID string
	OrgID    string
	UserID   string
	Role     string
	IsAdmin  bool
}

// AuthorizerACL is the narrow ACL surface the Authorizer depends on: per-row
// visibility evaluation plus group-membership resolution. The concrete ACL
// repository satisfies it via its segregated role ports.
type AuthorizerACL interface {
	repository.DocVisibilityPort
	repository.FolderVisibilityPort
	repository.GroupMembershipPort
}

// Authorizer is the application's row-level access decision point. It answers
// per-resource read/write visibility, applies the admin short-circuit, and
// resolves the caller's group memberships. The 404-not-403 invariant lives
// here: an invisible resource OR any lookup error maps to a NotFound apperr, so
// callers can never distinguish "exists but hidden" from "absent" and cannot
// probe for the existence of resources they may not see. The transport layer
// renders that NotFound as HTTP 404.
type Authorizer struct {
	acl AuthorizerACL
}

// NewAuthorizer builds an Authorizer over the given ACL surface.
func NewAuthorizer(acl AuthorizerACL) *Authorizer {
	return &Authorizer{acl: acl}
}

// notFound is the single, deliberately uniform "not visible" result. Document
// and folder messages differ only so existing client copy is preserved; both
// are KindNotFound with the NOT_FOUND code.
func docNotFound() error    { return apperr.NewNotFound("NOT_FOUND", "document not found") }
func folderNotFound() error { return apperr.NewNotFound("NOT_FOUND", "folder not found") }

// RequireDocVisible returns nil when the principal may read the document, or a
// NotFound error otherwise. Admins short-circuit without any DB lookup.
func (a *Authorizer) RequireDocVisible(ctx context.Context, p Principal, documentID string) error {
	if p.IsAdmin {
		return nil
	}
	groupIDs, err := a.acl.ListUserGroupIDs(ctx, p.UserID, p.OrgID)
	if err != nil {
		return docNotFound()
	}
	visible, err := a.acl.IsDocumentVisible(ctx, repository.VisibilityParams{
		TenantID:   p.TenantID,
		OrgID:      p.OrgID,
		DocumentID: documentID,
		UserID:     p.UserID,
		GroupIDs:   groupIDs,
		IsAdmin:    false,
	})
	if err != nil || !visible {
		return docNotFound()
	}
	return nil
}

// RequireDocWritable returns nil when the principal may write (edit/move/rename)
// the document, or a NotFound error otherwise. Org-open documents remain
// writable by role; a restricted document requires a write/delete grant. Admins
// short-circuit without any DB lookup. A non-writable document is reported as
// NotFound (not Forbidden) so a read-only grant cannot be used to probe for, or
// mutate, a restricted document.
func (a *Authorizer) RequireDocWritable(ctx context.Context, p Principal, documentID string) error {
	if p.IsAdmin {
		return nil
	}
	groupIDs, err := a.acl.ListUserGroupIDs(ctx, p.UserID, p.OrgID)
	if err != nil {
		return docNotFound()
	}
	writable, err := a.acl.IsDocumentWritable(ctx, repository.VisibilityParams{
		TenantID:   p.TenantID,
		OrgID:      p.OrgID,
		DocumentID: documentID,
		UserID:     p.UserID,
		GroupIDs:   groupIDs,
		IsAdmin:    false,
	})
	if err != nil || !writable {
		return docNotFound()
	}
	return nil
}

// RequireFolderVisible returns nil when the principal may read the folder, or a
// NotFound error otherwise. Admins short-circuit without any DB lookup.
func (a *Authorizer) RequireFolderVisible(ctx context.Context, p Principal, folderID string) error {
	if p.IsAdmin {
		return nil
	}
	groupIDs, err := a.acl.ListUserGroupIDs(ctx, p.UserID, p.OrgID)
	if err != nil {
		return folderNotFound()
	}
	visible, err := a.acl.IsFolderVisible(ctx, repository.FolderVisibilityParams{
		TenantID: p.TenantID,
		OrgID:    p.OrgID,
		FolderID: folderID,
		UserID:   p.UserID,
		GroupIDs: groupIDs,
		IsAdmin:  false,
	})
	if err != nil || !visible {
		return folderNotFound()
	}
	return nil
}
