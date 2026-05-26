package repository

import (
	"context"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
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
