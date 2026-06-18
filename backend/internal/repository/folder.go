package repository

import (
	"context"
	"errors"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errFolderNameExists is the unexported sentinel exposed via ErrFolderNameExists.
var errFolderNameExists = errors.New("folder name already exists under parent")

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type folderRepository struct {
	queries sqldb.Querier
}

func NewFolderRepository(db *pgxpool.Pool) FolderRepository {
	return &folderRepository{queries: sqldb.New(db)}
}

func (r *folderRepository) Create(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	err := r.queries.CreateFolder(ctx, sqldb.CreateFolderParams{
		ID:        folder.ID,
		TenantID:  folder.TenantID,
		OrgID:     folder.OrgID,
		ParentID:  folder.ParentID,
		Name:      folder.Name,
		CreatedBy: folder.CreatedBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return errFolderNameExists
		}
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (r *folderRepository) GetByParentName(ctx context.Context, tenantID, orgID string, parentID *string, name string) (*model.Folder, error) {
	folder, err := r.queries.GetFolderByParentName(ctx, sqldb.GetFolderByParentNameParams{
		TenantID: tenantID,
		OrgID:    orgID,
		ParentID: parentID,
		Name:     name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder by name: %w", err)
	}
	modelFolder := toModelFolder(folder)
	return &modelFolder, nil
}

func (r *folderRepository) GetAncestorIDs(ctx context.Context, tenantID, orgID, folderID string) ([]string, error) {
	ids, err := r.queries.GetFolderAncestorIDs(ctx, sqldb.GetFolderAncestorIDsParams{
		FolderID: folderID,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get folder ancestors: %w", err)
	}
	return ids, nil
}

// Move reparents a folder atomically. The cycle check lives inside the single
// MoveFolder UPDATE: the row is updated only when the new parent is neither the
// folder itself nor any of its descendants. A 0 rows-affected result therefore
// signals a rejected (cycle/invalid-parent) move and is surfaced to the caller
// via the returned row count rather than as a database error.
func (r *folderRepository) Move(ctx context.Context, tenantID, orgID, folderID string, parentID *string) (int64, error) {
	rows, err := r.queries.MoveFolder(ctx, sqldb.MoveFolderParams{
		NewParent: parentID,
		ID:        folderID,
		TenantID:  tenantID,
		OrgID:     orgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to move folder: %w", err)
	}
	return rows, nil
}

// SubtreeHeight returns the height of the subtree rooted at folderID: the
// number of folder levels from the folder itself down to its deepest
// descendant. The folder alone has height 1. A folder that does not exist
// returns height 0.
func (r *folderRepository) SubtreeHeight(ctx context.Context, tenantID, orgID, folderID string) (int, error) {
	height, err := r.queries.GetFolderSubtreeHeight(ctx, sqldb.GetFolderSubtreeHeightParams{
		FolderID: folderID,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get folder subtree height: %w", err)
	}
	return int(height), nil
}

func (r *folderRepository) GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Folder, error) {
	folder, err := r.queries.GetFolderByID(ctx, sqldb.GetFolderByIDParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	modelFolder := toModelFolder(folder)
	return &modelFolder, nil
}

func (r *folderRepository) ListByParent(ctx context.Context, tenantID, orgID, parentID string) ([]model.Folder, error) {
	folders, err := r.queries.ListFoldersByParent(ctx, sqldb.ListFoldersByParentParams{
		TenantID: tenantID,
		OrgID:    orgID,
		ParentID: &parentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	return toModelFolders(folders), nil
}

func (r *folderRepository) ListRoot(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	folders, err := r.queries.ListRootFolders(ctx, sqldb.ListRootFoldersParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list root folders: %w", err)
	}
	return toModelFolders(folders), nil
}

func (r *folderRepository) ListAll(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	folders, err := r.queries.ListAllFolders(ctx, sqldb.ListAllFoldersParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all folders: %w", err)
	}
	return toModelFolders(folders), nil
}

func (r *folderRepository) Update(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	err := r.queries.UpdateFolder(ctx, sqldb.UpdateFolderParams{
		Name:     folder.Name,
		ParentID: folder.ParentID,
		ID:       folder.ID,
		TenantID: folder.TenantID,
		OrgID:    folder.OrgID,
	})
	if err != nil {
		return fmt.Errorf("failed to update folder: %w", err)
	}
	return nil
}

func (r *folderRepository) Delete(ctx context.Context, tenantID, orgID, id string) error {
	rowsAffected, err := r.queries.DeleteFolder(ctx, sqldb.DeleteFolderParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}

func toModelFolders(folders []sqldb.Folder) []model.Folder {
	models := make([]model.Folder, 0, len(folders))
	for _, folder := range folders {
		models = append(models, toModelFolder(folder))
	}
	return models
}

func toModelFolder(folder sqldb.Folder) model.Folder {
	return model.Folder{
		ID:        folder.ID,
		TenantID:  folder.TenantID,
		OrgID:     folder.OrgID,
		ParentID:  folder.ParentID,
		Name:      folder.Name,
		CreatedBy: folder.CreatedBy,
		CreatedAt: folder.CreatedAt.Time,
	}
}
