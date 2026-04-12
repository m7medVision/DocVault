package service

import (
	"context"
	"fmt"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

type FolderService struct {
	repo repository.FolderRepository
}

func NewFolderService(repo repository.FolderRepository) *FolderService {
	return &FolderService{repo: repo}
}

type CreateFolderInput struct {
	TenantID  string
	OrgID     string
	ParentID  *string
	Name      string
	CreatedBy string
}

type CreateFolderOutput struct {
	Folder model.Folder
}

type ListFoldersInput struct {
	TenantID string
	OrgID    string
	ParentID *string
}

type ListFoldersOutput struct {
	Folders []model.Folder
}

func (s *FolderService) Create(ctx context.Context, input *CreateFolderInput) (*CreateFolderOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("folder name is required")
	}

	if s.repo == nil {
		return nil, ErrFolderRepositoryNotConfigured
	}

	if input.ParentID != nil && *input.ParentID != "" {
		_, err := s.repo.GetByID(ctx, input.TenantID, input.OrgID, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent folder not found: %w", err)
		}
	}

	folder := model.Folder{
		ID:        uuid.New().String(),
		TenantID:  input.TenantID,
		OrgID:     input.OrgID,
		ParentID:  input.ParentID,
		Name:      input.Name,
		CreatedBy: &input.CreatedBy,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, &folder); err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return &CreateFolderOutput{Folder: folder}, nil
}

func (s *FolderService) List(ctx context.Context, input *ListFoldersInput) (*ListFoldersOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}

	if s.repo == nil {
		return nil, ErrFolderRepositoryNotConfigured
	}

	var folders []model.Folder
	var err error

	if input.ParentID != nil {
		folders, err = s.repo.ListByParent(ctx, input.TenantID, input.OrgID, *input.ParentID)
	} else {
		folders, err = s.repo.ListRoot(ctx, input.TenantID, input.OrgID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	return &ListFoldersOutput{Folders: folders}, nil
}
