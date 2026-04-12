package repository

import (
	"context"
	"fmt"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	if tag == nil {
		return fmt.Errorf("tag is nil")
	}
	query := `
		INSERT INTO tags (id, tenant_id, name, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, name) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, tag.ID, tag.TenantID, tag.Name)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func (r *tagRepository) GetByID(ctx context.Context, tenantID, id string) (*model.Tag, error) {
	query := `SELECT id, tenant_id, name, created_at FROM tags WHERE id = $1 AND tenant_id = $2`
	var tag model.Tag
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return &tag, nil
}

func (r *tagRepository) GetByName(ctx context.Context, tenantID, name string) (*model.Tag, error) {
	query := `SELECT id, tenant_id, name, created_at FROM tags WHERE tenant_id = $1 AND name = $2`
	var tag model.Tag
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag by name: %w", err)
	}
	return &tag, nil
}

func (r *tagRepository) List(ctx context.Context, tenantID string, query string, limit int) ([]model.Tag, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	sql := `SELECT id, tenant_id, name, created_at FROM tags WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1

	if query != "" {
		argCount++
		sql += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+query+"%")
	}

	sql += fmt.Sprintf(" ORDER BY name ASC LIMIT %d", limit)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *tagRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM tags WHERE id = $1 AND tenant_id = $2`
	result, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func (r *tagRepository) AddToDocument(ctx context.Context, tenantID, tagID, documentID string) error {
	query := `
		INSERT INTO document_tags (document_id, tag_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, documentID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to document: %w", err)
	}
	return nil
}

func (r *tagRepository) RemoveFromDocument(ctx context.Context, tenantID, tagID, documentID string) error {
	query := `DELETE FROM document_tags WHERE document_id = $1 AND tag_id = $2`
	_, err := r.db.Exec(ctx, query, documentID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from document: %w", err)
	}
	return nil
}

func (r *tagRepository) GetDocumentTags(ctx context.Context, tenantID, documentID string) ([]model.Tag, error) {
	query := `
		SELECT t.id, t.tenant_id, t.name, t.created_at
		FROM tags t
		JOIN document_tags dt ON t.id = dt.tag_id
		WHERE dt.document_id = $1 AND t.tenant_id = $2
		ORDER BY t.name ASC
	`
	rows, err := r.db.Query(ctx, query, documentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}
