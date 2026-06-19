package postgres

import (
	"context"
	"errors"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type aclRepository struct {
	queries sqldb.Querier
}

// NewACLRepository creates a new ACLRepository.
func NewACLRepository(db *pgxpool.Pool) repository.ACLRepository {
	return &aclRepository{queries: sqldb.New(db)}
}

// IsDocumentVisible reports whether the principal may read the document.
// A missing document maps to (false, nil) so callers cannot distinguish
// "not found" from "not visible" (avoids leaking existence).
func (r *aclRepository) IsDocumentVisible(ctx context.Context, params repository.VisibilityParams) (bool, error) {
	groupIDs := params.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}

	visible, err := r.queries.IsDocumentVisibleToUser(ctx, sqldb.IsDocumentVisibleToUserParams{
		IsAdmin:    params.IsAdmin,
		UserID:     params.UserID,
		GroupIds:   groupIDs,
		DocumentID: params.DocumentID,
		TenantID:   params.TenantID,
		OrgID:      params.OrgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to evaluate document visibility: %w", err)
	}
	if visible == nil {
		return false, nil
	}
	return *visible, nil
}

// IsDocumentWritable reports whether the principal may write (edit/move/rename)
// the document. Org-open documents remain writable by role; restricted documents
// require a 'write' or 'delete' grant. A missing document maps to (false, nil) so
// callers cannot distinguish "not found" from "not writable" (avoids leaking
// existence).
func (r *aclRepository) IsDocumentWritable(ctx context.Context, params repository.VisibilityParams) (bool, error) {
	groupIDs := params.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}

	writable, err := r.queries.IsDocumentWritableToUser(ctx, sqldb.IsDocumentWritableToUserParams{
		IsAdmin:    params.IsAdmin,
		UserID:     params.UserID,
		GroupIds:   groupIDs,
		DocumentID: params.DocumentID,
		TenantID:   params.TenantID,
		OrgID:      params.OrgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to evaluate document writability: %w", err)
	}
	if writable == nil {
		return false, nil
	}
	return *writable, nil
}

// IsFolderVisible reports whether the principal may read the folder. A missing
// folder maps to (false, nil) so callers cannot distinguish "not found" from
// "not visible" (avoids leaking existence).
func (r *aclRepository) IsFolderVisible(ctx context.Context, params repository.FolderVisibilityParams) (bool, error) {
	groupIDs := params.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}

	visible, err := r.queries.IsFolderVisibleToUser(ctx, sqldb.IsFolderVisibleToUserParams{
		IsAdmin:  params.IsAdmin,
		UserID:   params.UserID,
		GroupIds: groupIDs,
		FolderID: params.FolderID,
		TenantID: params.TenantID,
		OrgID:    params.OrgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to evaluate folder visibility: %w", err)
	}
	if visible == nil {
		return false, nil
	}
	return *visible, nil
}

// ListUserGroupIDs returns the group IDs the user belongs to within the org.
func (r *aclRepository) ListUserGroupIDs(ctx context.Context, userID, orgID string) ([]string, error) {
	groupIDs, err := r.queries.ListGroupIDsForUser(ctx, sqldb.ListGroupIDsForUserParams{
		UserID: userID,
		OrgID:  orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list user group ids: %w", err)
	}
	return groupIDs, nil
}

// ListVisibleDocuments lists documents the principal may read, applying the
// optional doc_type/folder/status/language filters and keyset (cursor)
// pagination. It returns the next-page cursor when more rows remain.
func (r *aclRepository) ListVisibleDocuments(ctx context.Context, params repository.ListVisibleParams) ([]model.Document, *string, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	groupIDs := params.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}

	rows, err := r.queries.ListVisibleDocuments(ctx, sqldb.ListVisibleDocumentsParams{
		TenantID:   params.TenantID,
		OrgID:      params.OrgID,
		DocType:    optionalString(params.DocType),
		FolderID:   optionalString(params.FolderID),
		Status:     optionalString(params.Status),
		Language:   optionalString(params.Language),
		CursorID:   optionalString(params.Cursor),
		IsAdmin:    params.IsAdmin,
		UserID:     params.UserID,
		GroupIds:   groupIDs,
		LimitCount: int32(limit + 1),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list visible documents: %w", err)
	}

	docs := make([]model.Document, 0, len(rows))
	for _, row := range rows {
		docs = append(docs, model.Document{
			ID:        row.ID,
			TenantID:  row.TenantID,
			OrgID:     row.OrgID,
			FolderID:  row.FolderID,
			OwnerID:   row.OwnerID,
			Title:     row.Title,
			DocType:   string(row.DocType),
			Status:    model.DocumentStatus(row.Status),
			Language:  row.Language,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	var cursor *string
	if len(docs) > limit {
		docs = docs[:limit]
		c := docs[len(docs)-1].ID
		cursor = &c
	}

	return docs, cursor, nil
}

// CreateGroup creates an org-scoped group and returns the persisted row.
func (r *aclRepository) CreateGroup(ctx context.Context, params repository.CreateGroupParams) (repository.Group, error) {
	row, err := r.queries.CreateGroup(ctx, sqldb.CreateGroupParams{
		TenantID:  params.TenantID,
		OrgID:     params.OrgID,
		Name:      params.Name,
		CreatedBy: params.CreatedBy,
	})
	if err != nil {
		return repository.Group{}, fmt.Errorf("failed to create group: %w", err)
	}
	return repository.Group{
		ID:        row.ID,
		TenantID:  row.TenantID,
		OrgID:     row.OrgID,
		Name:      row.Name,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

// DeleteGroup deletes an org-scoped group and returns the affected row count.
func (r *aclRepository) DeleteGroup(ctx context.Context, tenantID, orgID, groupID string) (int64, error) {
	rows, err := r.queries.DeleteGroup(ctx, sqldb.DeleteGroupParams{
		ID:       groupID,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to delete group: %w", err)
	}
	return rows, nil
}

// ListGroups returns the groups defined within the org.
func (r *aclRepository) ListGroups(ctx context.Context, tenantID, orgID string) ([]repository.Group, error) {
	rows, err := r.queries.ListGroups(ctx, sqldb.ListGroupsParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	groups := make([]repository.Group, 0, len(rows))
	for _, row := range rows {
		groups = append(groups, repository.Group{
			ID:        row.ID,
			TenantID:  row.TenantID,
			OrgID:     row.OrgID,
			Name:      row.Name,
			CreatedBy: row.CreatedBy,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return groups, nil
}

// AddGroupMember adds a user to a group (idempotent; org-scoped via the group).
func (r *aclRepository) AddGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) error {
	if err := r.queries.AddGroupMember(ctx, sqldb.AddGroupMemberParams{
		GroupID:  groupID,
		UserID:   userID,
		TenantID: tenantID,
		OrgID:    orgID,
	}); err != nil {
		return fmt.Errorf("failed to add group member: %w", err)
	}
	return nil
}

// RemoveGroupMember removes a user from a group and returns the affected count.
func (r *aclRepository) RemoveGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) (int64, error) {
	rows, err := r.queries.RemoveGroupMember(ctx, sqldb.RemoveGroupMemberParams{
		GroupID:  groupID,
		UserID:   userID,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to remove group member: %w", err)
	}
	return rows, nil
}

// CreateGrant creates a per-resource grant (idempotent) and returns its id.
// An empty id with a nil error means the grant already existed.
func (r *aclRepository) CreateGrant(ctx context.Context, params repository.CreateGrantParams) (string, error) {
	id, err := r.queries.CreateGrant(ctx, sqldb.CreateGrantParams{
		TenantID:      params.TenantID,
		OrgID:         params.OrgID,
		ResourceType:  sqldb.AclResourceType(params.ResourceType),
		ResourceID:    params.ResourceID,
		PrincipalType: sqldb.AclPrincipalType(params.PrincipalType),
		PrincipalID:   params.PrincipalID,
		Permission:    sqldb.AclPermission(params.Permission),
		GrantedBy:     params.GrantedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to create grant: %w", err)
	}
	return id, nil
}

// GrantTargetExists reports whether the grant target (a document or folder
// identified by resource_type/resource_id) exists within the caller's org.
func (r *aclRepository) GrantTargetExists(ctx context.Context, ref repository.ResourceRef) (bool, error) {
	exists, err := r.queries.GrantTargetExists(ctx, sqldb.GrantTargetExistsParams{
		ResourceType: sqldb.AclResourceType(ref.ResourceType),
		ResourceID:   ref.ResourceID,
		TenantID:     ref.TenantID,
		OrgID:        ref.OrgID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check grant target existence: %w", err)
	}
	return exists, nil
}

// DeleteGrant deletes a grant by id and returns the affected row count.
func (r *aclRepository) DeleteGrant(ctx context.Context, tenantID, orgID, grantID string) (int64, error) {
	rows, err := r.queries.DeleteGrant(ctx, sqldb.DeleteGrantParams{
		ID:       grantID,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to delete grant: %w", err)
	}
	return rows, nil
}

// ListGrantsByResource lists the grants attached to a single resource.
func (r *aclRepository) ListGrantsByResource(ctx context.Context, ref repository.ResourceRef) ([]repository.Grant, error) {
	rows, err := r.queries.ListGrantsByResource(ctx, sqldb.ListGrantsByResourceParams{
		TenantID:     ref.TenantID,
		OrgID:        ref.OrgID,
		ResourceType: sqldb.AclResourceType(ref.ResourceType),
		ResourceID:   ref.ResourceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}
	grants := make([]repository.Grant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, repository.Grant{
			ID:            row.ID,
			TenantID:      row.TenantID,
			OrgID:         row.OrgID,
			ResourceType:  string(row.ResourceType),
			ResourceID:    row.ResourceID,
			PrincipalType: string(row.PrincipalType),
			PrincipalID:   row.PrincipalID,
			Permission:    string(row.Permission),
			GrantedBy:     row.GrantedBy,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return grants, nil
}

// DeleteGrantsForResource removes every grant attached to a resource (used on
// resource deletion) and returns the affected row count.
func (r *aclRepository) DeleteGrantsForResource(ctx context.Context, ref repository.ResourceRef) (int64, error) {
	rows, err := r.queries.DeleteGrantsForResource(ctx, sqldb.DeleteGrantsForResourceParams{
		TenantID:     ref.TenantID,
		OrgID:        ref.OrgID,
		ResourceType: sqldb.AclResourceType(ref.ResourceType),
		ResourceID:   ref.ResourceID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to delete grants for resource: %w", err)
	}
	return rows, nil
}

// SetDocumentRestricted toggles a document's is_restricted flag.
func (r *aclRepository) SetDocumentRestricted(ctx context.Context, params repository.SetRestrictedParams) (int64, error) {
	rows, err := r.queries.SetDocumentRestricted(ctx, sqldb.SetDocumentRestrictedParams{
		IsRestricted: params.Restricted,
		ID:           params.ResourceID,
		TenantID:     params.TenantID,
		OrgID:        params.OrgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to set document restriction: %w", err)
	}
	return rows, nil
}

// SetFolderRestricted toggles a folder's is_restricted flag.
func (r *aclRepository) SetFolderRestricted(ctx context.Context, params repository.SetRestrictedParams) (int64, error) {
	rows, err := r.queries.SetFolderRestricted(ctx, sqldb.SetFolderRestrictedParams{
		IsRestricted: params.Restricted,
		ID:           params.ResourceID,
		TenantID:     params.TenantID,
		OrgID:        params.OrgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to set folder restriction: %w", err)
	}
	return rows, nil
}
