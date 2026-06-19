package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/platform/cache"
	"github.com/docvault/backend/internal/repository"
)

// folderTreeTTL bounds staleness of the cached org folder tree. Structural
// mutations (create/rename/move/delete) bust the key explicitly, so the TTL
// only ever covers an is_restricted toggle made through the ACL repository —
// which is display-only here (access is enforced by the visibility queries, not
// this list), so a short bound is fine.
const folderTreeTTL = 60 * time.Second

// CachingFolders decorates a FolderRepository with a short-lived cache for the
// per-org ListAll folder tree, which is read on every page that renders the
// navigation tree. The result is org-uniform (no per-user filtering), so a
// single per-(tenant,org) entry serves all members. Structural mutations bust
// the entry; everything else is delegated unchanged.
type CachingFolders struct {
	repository.FolderRepository
	cache cache.Cache
}

// NewCachingFolders wraps inner so ListAll reads through cache.
func NewCachingFolders(inner repository.FolderRepository, c cache.Cache) *CachingFolders {
	return &CachingFolders{FolderRepository: inner, cache: c}
}

func folderTreeKey(tenantID, orgID string) string {
	return fmt.Sprintf("folders:all:%s:%s", tenantID, orgID)
}

// ListAll returns the cached folder tree when present, otherwise reads through
// to the inner repository and caches it. A cache error falls through to the
// authoritative query (best-effort).
func (c *CachingFolders) ListAll(ctx context.Context, tenantID, orgID string) ([]document.Folder, error) {
	key := folderTreeKey(tenantID, orgID)
	if raw, ok, err := c.cache.Get(ctx, key); err == nil && ok {
		var folders []document.Folder
		if json.Unmarshal(raw, &folders) == nil {
			return folders, nil
		}
	}

	folders, err := c.FolderRepository.ListAll(ctx, tenantID, orgID)
	if err != nil {
		return nil, err
	}
	if raw, mErr := json.Marshal(folders); mErr == nil {
		if sErr := c.cache.Set(ctx, key, raw, folderTreeTTL); sErr != nil {
			slog.WarnContext(ctx, "folder tree cache set failed", "error", sErr)
		}
	}
	return folders, nil
}

// Create creates the folder, then busts its org's cached tree.
func (c *CachingFolders) Create(ctx context.Context, folder *document.Folder) error {
	if err := c.FolderRepository.Create(ctx, folder); err != nil {
		return err
	}
	c.bustTree(ctx, folder.TenantID, folder.OrgID)
	return nil
}

// Update renames the folder, then busts its org's cached tree.
func (c *CachingFolders) Update(ctx context.Context, folder *document.Folder) error {
	if err := c.FolderRepository.Update(ctx, folder); err != nil {
		return err
	}
	c.bustTree(ctx, folder.TenantID, folder.OrgID)
	return nil
}

// Reparent moves the folder, then busts its org's cached tree.
func (c *CachingFolders) Reparent(ctx context.Context, tenantID, orgID, folderID string, parentID *string, maxDepth int) error {
	if err := c.FolderRepository.Reparent(ctx, tenantID, orgID, folderID, parentID, maxDepth); err != nil {
		return err
	}
	c.bustTree(ctx, tenantID, orgID)
	return nil
}

// Delete deletes the folder, then busts its org's cached tree.
func (c *CachingFolders) Delete(ctx context.Context, tenantID, orgID, id string) error {
	if err := c.FolderRepository.Delete(ctx, tenantID, orgID, id); err != nil {
		return err
	}
	c.bustTree(ctx, tenantID, orgID)
	return nil
}

func (c *CachingFolders) bustTree(ctx context.Context, tenantID, orgID string) {
	if err := c.cache.Del(ctx, folderTreeKey(tenantID, orgID)); err != nil {
		slog.WarnContext(ctx, "folder tree cache bust failed", "error", err)
	}
}
