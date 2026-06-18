package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	model "github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

// SplitFolderPath splits a nested folder path on "/", strips surrounding
// whitespace from each segment, and skips empty segments. It implements the
// nested-path contract shared with the processing producer: the
// documents.suggested_folder_name column stores segments joined by "/".
func SplitFolderPath(path string) []string {
	parts := strings.Split(path, "/")
	segments := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		segments = append(segments, trimmed)
	}
	return segments
}

type FolderService struct {
	repo    repository.FolderRepository
	aclRepo repository.ACLRepository
}

func NewFolderService(repo repository.FolderRepository, aclRepo repository.ACLRepository) *FolderService {
	return &FolderService{repo: repo, aclRepo: aclRepo}
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

func (s *FolderService) ListAll(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}

	if s.repo == nil {
		return nil, ErrFolderRepositoryNotConfigured
	}

	return s.repo.ListAll(ctx, tenantID, orgID)
}

func (s *FolderService) Rename(ctx context.Context, tenantID, orgID, folderID, newName string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if orgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if folderID == "" {
		return fmt.Errorf("folder_id is required")
	}
	if newName == "" {
		return fmt.Errorf("new_name is required")
	}

	folder, err := s.repo.GetByID(ctx, tenantID, orgID, folderID)
	if err != nil {
		return fmt.Errorf("folder not found: %w", err)
	}

	folder.Name = newName
	if err := s.repo.Update(ctx, folder); err != nil {
		return fmt.Errorf("failed to rename folder: %w", err)
	}

	return nil
}

// EnsureFolderPath walks the given path segments top-down starting from the
// root (parent_id NULL) and find-or-creates each segment as a folder under the
// running parent, returning the leaf folder's id. Segments are expected to be
// already normalized (split on "/", trimmed, empty-skipped) per the nested-path
// contract; callers can use SplitFolderPath. A unique-name violation during
// create is treated as "already exists" and the existing folder is fetched
// instead, making the operation idempotent under concurrency.
func (s *FolderService) EnsureFolderPath(ctx context.Context, tenantID, orgID, userID string, segments []string) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if orgID == "" {
		return "", fmt.Errorf("org_id is required")
	}
	if s.repo == nil {
		return "", ErrFolderRepositoryNotConfigured
	}
	if len(segments) == 0 {
		return "", fmt.Errorf("folder path is empty")
	}

	var parentID *string
	for _, name := range segments {
		// Try to find an existing folder with this name under the running parent.
		existing, err := s.repo.GetByParentName(ctx, tenantID, orgID, parentID, name)
		if err == nil {
			id := existing.ID
			parentID = &id
			continue
		}
		if !strings.Contains(err.Error(), "not found") {
			return "", fmt.Errorf("failed to look up folder %q: %w", name, err)
		}

		// Not found: create it. Carry the parent down by value.
		var parentCopy *string
		if parentID != nil {
			p := *parentID
			parentCopy = &p
		}
		folder := model.Folder{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			OrgID:     orgID,
			ParentID:  parentCopy,
			Name:      name,
			CreatedBy: &userID,
			CreatedAt: time.Now(),
		}
		createErr := s.repo.Create(ctx, &folder)
		if createErr == nil {
			id := folder.ID
			parentID = &id
			continue
		}
		// Lost a race or it already existed: fetch the existing one.
		if errors.Is(createErr, repository.ErrFolderNameExists) {
			existing, getErr := s.repo.GetByParentName(ctx, tenantID, orgID, parentID, name)
			if getErr != nil {
				return "", fmt.Errorf("failed to fetch existing folder %q after conflict: %w", name, getErr)
			}
			id := existing.ID
			parentID = &id
			continue
		}
		return "", fmt.Errorf("failed to create folder %q: %w", name, createErr)
	}

	if parentID == nil {
		return "", fmt.Errorf("folder path resolved to no leaf")
	}
	return *parentID, nil
}

func (s *FolderService) Delete(ctx context.Context, tenantID, orgID, folderID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if orgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if folderID == "" {
		return fmt.Errorf("folder_id is required")
	}

	if err := s.repo.Delete(ctx, tenantID, orgID, folderID); err != nil {
		return err
	}

	if s.aclRepo != nil {
		if _, err := s.aclRepo.DeleteGrantsForResource(ctx, repository.ResourceRef{
			TenantID:     tenantID,
			OrgID:        orgID,
			ResourceType: "folder",
			ResourceID:   folderID,
		}); err != nil {
			return fmt.Errorf("failed to clean up folder grants: %w", err)
		}
	}

	return nil
}
