package repository

import (
	"context"
	"errors"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Limit    int
}

// ACLRepository evaluates per-document/folder read visibility.
type ACLRepository interface {
	IsDocumentVisible(ctx context.Context, params VisibilityParams) (bool, error)
	ListUserGroupIDs(ctx context.Context, userID, orgID string) ([]string, error)
	ListVisibleDocuments(ctx context.Context, params ListVisibleParams) ([]model.Document, error)
}

type aclRepository struct {
	queries sqldb.Querier
}

// NewACLRepository creates a new ACLRepository.
func NewACLRepository(db *pgxpool.Pool) ACLRepository {
	return &aclRepository{queries: sqldb.New(db)}
}

// IsDocumentVisible reports whether the principal may read the document.
// A missing document maps to (false, nil) so callers cannot distinguish
// "not found" from "not visible" (avoids leaking existence).
func (r *aclRepository) IsDocumentVisible(ctx context.Context, params VisibilityParams) (bool, error) {
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
// optional doc_type/folder/status filters.
func (r *aclRepository) ListVisibleDocuments(ctx context.Context, params ListVisibleParams) ([]model.Document, error) {
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
		IsAdmin:    params.IsAdmin,
		UserID:     params.UserID,
		GroupIds:   groupIDs,
		LimitCount: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list visible documents: %w", err)
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

	return docs, nil
}
