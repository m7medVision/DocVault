package postgres

import (
	"context"
	"testing"

	"github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/repository"
)

// fakeInnerFolders embeds the FolderRepository interface (nil) and overrides
// only the methods the caching decorator touches.
type fakeInnerFolders struct {
	repository.FolderRepository
	folders     []document.Folder
	listAllCall int
}

func (f *fakeInnerFolders) ListAll(_ context.Context, _, _ string) ([]document.Folder, error) {
	f.listAllCall++
	return f.folders, nil
}
func (f *fakeInnerFolders) Create(_ context.Context, _ *document.Folder) error { return nil }
func (f *fakeInnerFolders) Update(_ context.Context, _ *document.Folder) error { return nil }
func (f *fakeInnerFolders) Reparent(_ context.Context, _, _, _ string, _ *string, _ int) error {
	return nil
}
func (f *fakeInnerFolders) Delete(_ context.Context, _, _, _ string) error { return nil }

func TestCachingFolders_SecondListHitsCache(t *testing.T) {
	inner := &fakeInnerFolders{folders: []document.Folder{{ID: "f1", Name: "A"}, {ID: "f2", Name: "B"}}}
	repo := NewCachingFolders(inner, newMemCache())
	ctx := context.Background()

	first, err := repo.ListAll(ctx, "t1", "o1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := repo.ListAll(ctx, "t1", "o1")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if inner.listAllCall != 1 {
		t.Fatalf("inner ListAll called %d times, want 1 (second should hit cache)", inner.listAllCall)
	}
	if len(first) != 2 || len(second) != 2 || second[0].ID != "f1" || second[1].Name != "B" {
		t.Fatalf("folder mismatch: first=%v second=%v", first, second)
	}
}

func TestCachingFolders_MutationsInvalidateTree(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(r *CachingFolders, ctx context.Context) error
	}{
		{"create", func(r *CachingFolders, ctx context.Context) error {
			return r.Create(ctx, &document.Folder{TenantID: "t1", OrgID: "o1"})
		}},
		{"update", func(r *CachingFolders, ctx context.Context) error {
			return r.Update(ctx, &document.Folder{TenantID: "t1", OrgID: "o1"})
		}},
		{"reparent", func(r *CachingFolders, ctx context.Context) error {
			return r.Reparent(ctx, "t1", "o1", "f1", nil, 12)
		}},
		{"delete", func(r *CachingFolders, ctx context.Context) error {
			return r.Delete(ctx, "t1", "o1", "f1")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := &fakeInnerFolders{folders: []document.Folder{{ID: "f1"}}}
			repo := NewCachingFolders(inner, newMemCache())
			ctx := context.Background()

			_, _ = repo.ListAll(ctx, "t1", "o1") // populate
			if err := tc.mutate(repo, ctx); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			_, _ = repo.ListAll(ctx, "t1", "o1") // must re-read

			if inner.listAllCall != 2 {
				t.Fatalf("inner ListAll called %d times, want 2 (%s must invalidate the tree)", inner.listAllCall, tc.name)
			}
		})
	}
}
