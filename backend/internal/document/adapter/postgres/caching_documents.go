package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/platform/cache"
	"github.com/docvault/backend/internal/repository"
)

// documentStatsTTL bounds how long cached org stats may be stale. Kept short
// because some inputs change outside this process: the processing workers move
// documents through statuses (pending -> processed) with direct DB writes, and
// the "completed this week" window slides with wall-clock time. Document
// create/delete in this process bust the key explicitly; the TTL covers the
// rest, which is acceptable for an approximate dashboard metric.
const documentStatsTTL = 60 * time.Second

// CachingDocuments decorates a DocumentRepository with a short-lived cache for
// the per-org GetDocumentStats aggregate, which scans the org's documents (and
// joins versions for storage) on every dashboard load. Only GetStats is cached;
// document create and delete bust the org's stats key. Everything else is
// delegated unchanged.
type CachingDocuments struct {
	repository.DocumentRepository
	cache cache.Cache
}

// NewCachingDocuments wraps inner so GetStats reads through cache.
func NewCachingDocuments(inner repository.DocumentRepository, c cache.Cache) *CachingDocuments {
	return &CachingDocuments{DocumentRepository: inner, cache: c}
}

func statsCacheKey(tenantID, orgID string) string {
	return fmt.Sprintf("stats:%s:%s", tenantID, orgID)
}

// GetStats returns the cached stats when present, otherwise reads through to the
// inner repository and caches the result. A cache error falls through to the
// authoritative query (best-effort), so the cache can never turn a working
// request into a failing one.
func (c *CachingDocuments) GetStats(ctx context.Context, tenantID, orgID string) (*document.DocumentStats, error) {
	key := statsCacheKey(tenantID, orgID)
	if raw, ok, err := c.cache.Get(ctx, key); err == nil && ok {
		var stats document.DocumentStats
		if json.Unmarshal(raw, &stats) == nil {
			return &stats, nil
		}
	}

	stats, err := c.DocumentRepository.GetStats(ctx, tenantID, orgID)
	if err != nil {
		return nil, err
	}
	if stats != nil {
		if raw, mErr := json.Marshal(stats); mErr == nil {
			if sErr := c.cache.Set(ctx, key, raw, documentStatsTTL); sErr != nil {
				slog.WarnContext(ctx, "document stats cache set failed", "error", sErr)
			}
		}
	}
	return stats, nil
}

// Create creates the document, then busts its org's cached stats.
func (c *CachingDocuments) Create(ctx context.Context, doc *document.Document) error {
	if err := c.DocumentRepository.Create(ctx, doc); err != nil {
		return err
	}
	c.bustStats(ctx, doc.TenantID, doc.OrgID)
	return nil
}

// Delete deletes the document, then busts its org's cached stats.
func (c *CachingDocuments) Delete(ctx context.Context, tenantID, orgID, id, actorID string) error {
	if err := c.DocumentRepository.Delete(ctx, tenantID, orgID, id, actorID); err != nil {
		return err
	}
	c.bustStats(ctx, tenantID, orgID)
	return nil
}

func (c *CachingDocuments) bustStats(ctx context.Context, tenantID, orgID string) {
	if err := c.cache.Del(ctx, statsCacheKey(tenantID, orgID)); err != nil {
		slog.WarnContext(ctx, "document stats cache bust failed", "error", err)
	}
}
