package repository

import (
	"context"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tagRepository struct {
	queries sqldb.Querier
}

func NewTagRepository(db *pgxpool.Pool) TagRepository {
	return &tagRepository{queries: sqldb.New(db)}
}

func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	if tag == nil {
		return fmt.Errorf("tag is nil")
	}
	err := r.queries.CreateTag(ctx, sqldb.CreateTagParams{
		ID:       tag.ID,
		TenantID: tag.TenantID,
		Name:     tag.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func (r *tagRepository) GetByID(ctx context.Context, tenantID, id string) (*model.Tag, error) {
	tag, err := r.queries.GetTagByID(ctx, sqldb.GetTagByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	modelTag := toModelTag(tag)
	return &modelTag, nil
}

func (r *tagRepository) GetByName(ctx context.Context, tenantID, name string) (*model.Tag, error) {
	tag, err := r.queries.GetTagByName(ctx, sqldb.GetTagByNameParams{
		TenantID: tenantID,
		Name:     name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag by name: %w", err)
	}
	modelTag := toModelTag(tag)
	return &modelTag, nil
}

func (r *tagRepository) List(ctx context.Context, tenantID string, query string, limit int) ([]model.Tag, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var (
		tags []sqldb.Tag
		err  error
	)
	if query != "" {
		tags, err = r.queries.SearchTags(ctx, sqldb.SearchTagsParams{
			TenantID: tenantID,
			Name:     "%" + query + "%",
			Limit:    int32(limit),
		})
	} else {
		tags, err = r.queries.ListTags(ctx, sqldb.ListTagsParams{
			TenantID: tenantID,
			Limit:    int32(limit),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return toModelTags(tags), nil
}

func (r *tagRepository) Delete(ctx context.Context, tenantID, id string) error {
	rowsAffected, err := r.queries.DeleteTag(ctx, sqldb.DeleteTagParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func (r *tagRepository) AddToDocument(ctx context.Context, tenantID, tagID, documentID string) error {
	err := r.queries.AddTagToDocument(ctx, sqldb.AddTagToDocumentParams{
		DocumentID: documentID,
		TagID:      tagID,
	})
	if err != nil {
		return fmt.Errorf("failed to add tag to document: %w", err)
	}
	return nil
}

func (r *tagRepository) RemoveFromDocument(ctx context.Context, tenantID, tagID, documentID string) error {
	err := r.queries.RemoveTagFromDocument(ctx, sqldb.RemoveTagFromDocumentParams{
		DocumentID: documentID,
		TagID:      tagID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove tag from document: %w", err)
	}
	return nil
}

func (r *tagRepository) GetDocumentTags(ctx context.Context, tenantID, documentID string) ([]model.Tag, error) {
	tags, err := r.queries.GetDocumentTags(ctx, sqldb.GetDocumentTagsParams{
		DocumentID: documentID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document tags: %w", err)
	}
	return toModelTags(tags), nil
}

func toModelTags(tags []sqldb.Tag) []model.Tag {
	models := make([]model.Tag, 0, len(tags))
	for _, tag := range tags {
		models = append(models, toModelTag(tag))
	}
	return models
}

func toModelTag(tag sqldb.Tag) model.Tag {
	return model.Tag{
		ID:        tag.ID,
		TenantID:  tag.TenantID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt.Time,
	}
}
