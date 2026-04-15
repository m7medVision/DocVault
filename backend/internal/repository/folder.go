package repository

import (
	"context"
	"fmt"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type folderRepository struct {
	db *pgxpool.Pool
}

func NewFolderRepository(db *pgxpool.Pool) FolderRepository {
	return &folderRepository{db: db}
}

func (r *folderRepository) Create(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	query := `
		INSERT INTO folders (id, tenant_id, org_id, parent_id, name, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		folder.ID, folder.TenantID, folder.OrgID,
		folder.ParentID, folder.Name, folder.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (r *folderRepository) GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Folder, error) {
	query := `
		SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
		FROM folders
		WHERE id = $1 AND tenant_id = $2 AND org_id = $3
	`
	var folder model.Folder
	err := r.db.QueryRow(ctx, query, id, tenantID, orgID).Scan(
		&folder.ID, &folder.TenantID, &folder.OrgID,
		&folder.ParentID, &folder.Name, &folder.CreatedBy, &folder.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	return &folder, nil
}

func (r *folderRepository) ListByParent(ctx context.Context, tenantID, orgID, parentID string) ([]model.Folder, error) {
	query := `
		SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
		FROM folders
		WHERE tenant_id = $1 AND org_id = $2 AND parent_id = $3
		ORDER BY name ASC
	`
	return r.queryFolders(ctx, query, tenantID, orgID, parentID)
}

func (r *folderRepository) ListRoot(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	query := `
		SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
		FROM folders
		WHERE tenant_id = $1 AND org_id = $2 AND parent_id IS NULL
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list root folders: %w", err)
	}
	defer rows.Close()
	return scanFolders(rows)
}

func (r *folderRepository) ListAll(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	query := `
		SELECT id, tenant_id, org_id, parent_id, name, created_by, created_at
		FROM folders
		WHERE tenant_id = $1 AND org_id = $2
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list all folders: %w", err)
	}
	defer rows.Close()
	return scanFolders(rows)
}

func (r *folderRepository) Update(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	query := `
		UPDATE folders SET name = $1, parent_id = $2
		WHERE id = $3 AND tenant_id = $4 AND org_id = $5
	`
	_, err := r.db.Exec(ctx, query,
		folder.Name, folder.ParentID, folder.ID, folder.TenantID, folder.OrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update folder: %w", err)
	}
	return nil
}

func (r *folderRepository) Delete(ctx context.Context, tenantID, orgID, id string) error {
	query := `DELETE FROM folders WHERE id = $1 AND tenant_id = $2 AND org_id = $3`
	result, err := r.db.Exec(ctx, query, id, tenantID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}

func (r *folderRepository) queryFolders(ctx context.Context, query string, args ...interface{}) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer rows.Close()
	return scanFolders(rows)
}

func scanFolders(rows pgx.Rows) ([]model.Folder, error) {
	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.OrgID, &f.ParentID,
			&f.Name, &f.CreatedBy, &f.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, f)
	}
	return folders, nil
}
