package app

import (
	"context"
	"fmt"

	model "github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/repository"
)

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(repo repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

type ListTagsInput struct {
	TenantID string
	Query    string
	Limit    int
}

type ListTagsOutput struct {
	Tags []model.Tag
}

func (s *TagService) List(ctx context.Context, input *ListTagsInput) (*ListTagsOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 50
	}

	if s.repo == nil {
		return nil, ErrTagRepositoryNotConfigured
	}

	tags, err := s.repo.List(ctx, input.TenantID, input.Query, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return &ListTagsOutput{Tags: tags}, nil
}
