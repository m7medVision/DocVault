package repository

import (
	"context"
	"testing"

	"github.com/docvault/backend/internal/domain/document"
)

// fakeInnerDocs embeds the DocumentRepository interface (nil) and overrides only
// the methods the caching decorator touches; any other call would panic, which
// is fine because the decorator delegates the rest verbatim.
type fakeInnerDocs struct {
	DocumentRepository
	stats      document.DocumentStats
	statsCalls int
}

func (f *fakeInnerDocs) GetStats(_ context.Context, _, _ string) (*document.DocumentStats, error) {
	f.statsCalls++
	s := f.stats
	return &s, nil
}
func (f *fakeInnerDocs) Create(_ context.Context, _ *document.Document) error { return nil }
func (f *fakeInnerDocs) Delete(_ context.Context, _, _, _, _ string) error    { return nil }

func TestCachingDocuments_SecondStatsCallHitsCache(t *testing.T) {
	inner := &fakeInnerDocs{stats: document.DocumentStats{TotalDocuments: 7, StorageUsedBytes: 1024}}
	repo := NewCachingDocuments(inner, newMemCache())
	ctx := context.Background()

	first, err := repo.GetStats(ctx, "t1", "o1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := repo.GetStats(ctx, "t1", "o1")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if inner.statsCalls != 1 {
		t.Fatalf("inner GetStats called %d times, want 1 (second should hit cache)", inner.statsCalls)
	}
	if first.TotalDocuments != 7 || second.TotalDocuments != 7 || second.StorageUsedBytes != 1024 {
		t.Fatalf("stats mismatch: first=%+v second=%+v", first, second)
	}
}

func TestCachingDocuments_DifferentOrgIsolated(t *testing.T) {
	inner := &fakeInnerDocs{stats: document.DocumentStats{TotalDocuments: 1}}
	repo := NewCachingDocuments(inner, newMemCache())
	ctx := context.Background()

	_, _ = repo.GetStats(ctx, "t1", "o1")
	_, _ = repo.GetStats(ctx, "t1", "o2") // different org -> separate key -> miss

	if inner.statsCalls != 2 {
		t.Fatalf("inner GetStats called %d times, want 2 (distinct orgs must not share a cache entry)", inner.statsCalls)
	}
}

func TestCachingDocuments_CreateInvalidatesStats(t *testing.T) {
	inner := &fakeInnerDocs{stats: document.DocumentStats{TotalDocuments: 1}}
	repo := NewCachingDocuments(inner, newMemCache())
	ctx := context.Background()

	_, _ = repo.GetStats(ctx, "t1", "o1") // populate
	if err := repo.Create(ctx, &document.Document{TenantID: "t1", OrgID: "o1"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, _ = repo.GetStats(ctx, "t1", "o1") // must re-read

	if inner.statsCalls != 2 {
		t.Fatalf("inner GetStats called %d times, want 2 (create must invalidate stats)", inner.statsCalls)
	}
}

func TestCachingDocuments_DeleteInvalidatesStats(t *testing.T) {
	inner := &fakeInnerDocs{stats: document.DocumentStats{TotalDocuments: 1}}
	repo := NewCachingDocuments(inner, newMemCache())
	ctx := context.Background()

	_, _ = repo.GetStats(ctx, "t1", "o1")
	if err := repo.Delete(ctx, "t1", "o1", "doc-1", "actor-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, _ = repo.GetStats(ctx, "t1", "o1")

	if inner.statsCalls != 2 {
		t.Fatalf("inner GetStats called %d times, want 2 (delete must invalidate stats)", inner.statsCalls)
	}
}
