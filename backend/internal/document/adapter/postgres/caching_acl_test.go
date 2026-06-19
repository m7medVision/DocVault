package postgres

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docvault/backend/internal/repository"
)

// fakeInnerACL embeds the ACLRepository interface (nil) and overrides only the
// group-membership methods the caching decorator touches; any other call would
// panic, which is fine because the decorator delegates the rest verbatim.
type fakeInnerACL struct {
	repository.ACLRepository
	groupIDs    []string
	listCalls   int
	addCalls    int
	removeCalls int
	deleteCalls int
}

func (f *fakeInnerACL) ListUserGroupIDs(_ context.Context, _, _ string) ([]string, error) {
	f.listCalls++
	return f.groupIDs, nil
}
func (f *fakeInnerACL) AddGroupMember(_ context.Context, _, _, _, _ string) error {
	f.addCalls++
	return nil
}
func (f *fakeInnerACL) RemoveGroupMember(_ context.Context, _, _, _, _ string) (int64, error) {
	f.removeCalls++
	return 1, nil
}
func (f *fakeInnerACL) DeleteGroup(_ context.Context, _, _, _ string) (int64, error) {
	f.deleteCalls++
	return 1, nil
}

// memCache is a minimal in-memory cache.Cache with prefix deletion.
type memCache struct {
	mu sync.Mutex
	m  map[string][]byte
}

func newMemCache() *memCache { return &memCache{m: map[string][]byte{}} }

func (c *memCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[key]
	return v, ok, nil
}
func (c *memCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return nil
}
func (c *memCache) Del(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.m, k)
	}
	return nil
}
func (c *memCache) DelByPrefix(_ context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.m {
		if strings.HasPrefix(k, prefix) {
			delete(c.m, k)
		}
	}
	return nil
}

func TestCachingACL_SecondLookupHitsCache(t *testing.T) {
	inner := &fakeInnerACL{groupIDs: []string{"g1", "g2"}}
	acl := NewCachingACL(inner, newMemCache())
	ctx := context.Background()

	first, err := acl.ListUserGroupIDs(ctx, "u1", "o1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := acl.ListUserGroupIDs(ctx, "u1", "o1")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if inner.listCalls != 1 {
		t.Fatalf("inner ListUserGroupIDs called %d times, want 1 (second should hit cache)", inner.listCalls)
	}
	if strings.Join(first, ",") != "g1,g2" || strings.Join(second, ",") != "g1,g2" {
		t.Fatalf("group ids mismatch: first=%v second=%v", first, second)
	}
}

func TestCachingACL_AddMemberInvalidatesThatUser(t *testing.T) {
	inner := &fakeInnerACL{groupIDs: []string{"g1"}}
	acl := NewCachingACL(inner, newMemCache())
	ctx := context.Background()

	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1") // populate
	if err := acl.AddGroupMember(ctx, "t1", "o1", "g9", "u1"); err != nil {
		t.Fatalf("add: %v", err)
	}
	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1") // must re-read

	if inner.listCalls != 2 {
		t.Fatalf("inner ListUserGroupIDs called %d times, want 2 (add must invalidate)", inner.listCalls)
	}
}

func TestCachingACL_RemoveMemberInvalidatesThatUser(t *testing.T) {
	inner := &fakeInnerACL{groupIDs: []string{"g1"}}
	acl := NewCachingACL(inner, newMemCache())
	ctx := context.Background()

	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1")
	if _, err := acl.RemoveGroupMember(ctx, "t1", "o1", "g1", "u1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1")

	if inner.listCalls != 2 {
		t.Fatalf("inner ListUserGroupIDs called %d times, want 2 (remove must invalidate)", inner.listCalls)
	}
}

func TestCachingACL_DeleteGroupInvalidatesWholeOrg(t *testing.T) {
	inner := &fakeInnerACL{groupIDs: []string{"g1"}}
	acl := NewCachingACL(inner, newMemCache())
	ctx := context.Background()

	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1") // listCalls=1
	_, _ = acl.ListUserGroupIDs(ctx, "u2", "o1") // listCalls=2
	if _, err := acl.DeleteGroup(ctx, "t1", "o1", "g1"); err != nil {
		t.Fatalf("delete group: %v", err)
	}
	_, _ = acl.ListUserGroupIDs(ctx, "u1", "o1") // must re-read -> 3
	_, _ = acl.ListUserGroupIDs(ctx, "u2", "o1") // must re-read -> 4

	if inner.listCalls != 4 {
		t.Fatalf("inner ListUserGroupIDs called %d times, want 4 (delete-group must invalidate both users)", inner.listCalls)
	}
}
