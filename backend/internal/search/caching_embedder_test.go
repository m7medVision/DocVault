package search

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakeCache is a minimal in-memory cache.Cache for tests.
type fakeCache struct {
	mu sync.Mutex
	m  map[string][]byte
}

func newFakeCache() *fakeCache { return &fakeCache{m: map[string][]byte{}} }

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[key]
	return v, ok, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return nil
}

func (c *fakeCache) Del(_ context.Context, _ ...string) error      { return nil }
func (c *fakeCache) DelByPrefix(_ context.Context, _ string) error { return nil }

type countingEmbedder struct {
	calls int
	vec   []float32
}

func (e *countingEmbedder) Embed(context.Context, string) ([]float32, error) {
	e.calls++
	return e.vec, nil
}

func (e *countingEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = e.vec
	}
	return out, nil
}

func TestCachingEmbedderHitsCacheOnRepeat(t *testing.T) {
	inner := &countingEmbedder{vec: []float32{0.1, -0.2, 3.5, 0}}
	ce := NewCachingEmbedder(inner, newFakeCache(), "model-x", 4, time.Hour)
	ctx := context.Background()

	first, err := ce.Embed(ctx, "hello")
	if err != nil {
		t.Fatalf("first embed: %v", err)
	}
	second, err := ce.Embed(ctx, "hello")
	if err != nil {
		t.Fatalf("second embed: %v", err)
	}

	if inner.calls != 1 {
		t.Fatalf("inner embedder called %d times, want 1 (second should hit cache)", inner.calls)
	}
	if len(first) != len(second) {
		t.Fatalf("length mismatch: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("cached vector differs at %d: %v vs %v", i, first[i], second[i])
		}
	}

	// A different text must miss and hit the embedder again.
	if _, err := ce.Embed(ctx, "world"); err != nil {
		t.Fatalf("third embed: %v", err)
	}
	if inner.calls != 2 {
		t.Fatalf("inner embedder called %d times, want 2", inner.calls)
	}
}

func TestVectorRoundTrip(t *testing.T) {
	vec := []float32{0, 1, -1, 3.14159, 2.5e-3}
	b, err := encodeVector(vec)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := decodeVector(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != len(vec) {
		t.Fatalf("length %d, want %d", len(got), len(vec))
	}
	for i := range vec {
		if got[i] != vec[i] {
			t.Fatalf("mismatch at %d: %v vs %v", i, got[i], vec[i])
		}
	}
}
