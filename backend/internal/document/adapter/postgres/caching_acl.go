package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/docvault/backend/internal/platform/cache"
	"github.com/docvault/backend/internal/repository"
)

// groupMembershipTTL bounds how long a cached (org,user) group set may be stale
// if an explicit invalidation is ever missed. Kept short because group
// membership feeds the visibility predicate: a removed member must lose access
// quickly. Explicit busting on every membership mutation is the primary
// mechanism; the TTL is the safety net.
const groupMembershipTTL = 60 * time.Second

// CachingACL decorates an ACLRepository with a short-lived cache for the
// per-(org,user) group-membership lookup, which the Authorizer performs on
// nearly every non-admin request. Only ListUserGroupIDs is cached; every
// membership mutation (add/remove member, delete group) busts the affected
// key(s) so a stale membership can never outlive the change beyond the TTL.
// Visibility evaluation and grants are delegated unchanged — they are never
// cached here, so restriction/grant changes take effect immediately.
type CachingACL struct {
	repository.ACLRepository
	cache cache.Cache
}

// NewCachingACL wraps inner so its group-membership lookups read through cache.
func NewCachingACL(inner repository.ACLRepository, c cache.Cache) *CachingACL {
	return &CachingACL{ACLRepository: inner, cache: c}
}

func groupsCacheKey(orgID, userID string) string {
	return fmt.Sprintf("groups:%s:%s", orgID, userID)
}

func groupsOrgPrefix(orgID string) string {
	return fmt.Sprintf("groups:%s:", orgID)
}

// ListUserGroupIDs returns the cached group set when present, otherwise reads
// through to the inner repository and caches the result. A cache error falls
// through to the authoritative lookup (best-effort), so the cache can never
// turn a working request into a failing one.
func (c *CachingACL) ListUserGroupIDs(ctx context.Context, userID, orgID string) ([]string, error) {
	key := groupsCacheKey(orgID, userID)
	if raw, ok, err := c.cache.Get(ctx, key); err == nil && ok {
		var ids []string
		if json.Unmarshal(raw, &ids) == nil {
			return ids, nil
		}
	}

	ids, err := c.ACLRepository.ListUserGroupIDs(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if raw, mErr := json.Marshal(ids); mErr == nil {
		if sErr := c.cache.Set(ctx, key, raw, groupMembershipTTL); sErr != nil {
			slog.WarnContext(ctx, "group membership cache set failed", "error", sErr)
		}
	}
	return ids, nil
}

// AddGroupMember adds the member, then busts that user's cached group set.
func (c *CachingACL) AddGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) error {
	if err := c.ACLRepository.AddGroupMember(ctx, tenantID, orgID, groupID, userID); err != nil {
		return err
	}
	c.bust(ctx, groupsCacheKey(orgID, userID))
	return nil
}

// RemoveGroupMember removes the member, then busts that user's cached group set.
func (c *CachingACL) RemoveGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) (int64, error) {
	rows, err := c.ACLRepository.RemoveGroupMember(ctx, tenantID, orgID, groupID, userID)
	if err != nil {
		return rows, err
	}
	c.bust(ctx, groupsCacheKey(orgID, userID))
	return rows, nil
}

// DeleteGroup deletes the group, then busts the whole org's group namespace:
// the group's removal drops memberships for any of its members, and the set of
// affected users is not known here.
func (c *CachingACL) DeleteGroup(ctx context.Context, tenantID, orgID, groupID string) (int64, error) {
	rows, err := c.ACLRepository.DeleteGroup(ctx, tenantID, orgID, groupID)
	if err != nil {
		return rows, err
	}
	if err := c.cache.DelByPrefix(ctx, groupsOrgPrefix(orgID)); err != nil {
		slog.WarnContext(ctx, "group membership cache prefix-bust failed", "error", err)
	}
	return rows, nil
}

func (c *CachingACL) bust(ctx context.Context, key string) {
	if err := c.cache.Del(ctx, key); err != nil {
		slog.WarnContext(ctx, "group membership cache bust failed", "error", err)
	}
}
