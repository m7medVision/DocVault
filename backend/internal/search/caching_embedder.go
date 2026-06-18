package search

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/docvault/backend/internal/platform/cache"
)

// CachingEmbedder decorates an Embedder with a content-addressed cache. Query
// embeddings are deterministic for a given (model, dimensions, text), so caching
// them removes a repeated external HTTP call on /search and /chat. The cache key
// is derived purely from the embedding content and carries no tenant data, so a
// single cached entry is safely shared across tenants.
//
// The cache is best-effort: any cache error falls through to the underlying
// embedder, so a Redis outage degrades performance but never correctness.
type CachingEmbedder struct {
	inner     Embedder
	cache     cache.Cache
	keyPrefix string
	ttl       time.Duration
}

var _ Embedder = (*CachingEmbedder)(nil)

// NewCachingEmbedder wraps inner with a cache keyed by (model, dimensions, text).
func NewCachingEmbedder(inner Embedder, c cache.Cache, model string, dimensions int, ttl time.Duration) *CachingEmbedder {
	return &CachingEmbedder{
		inner:     inner,
		cache:     c,
		keyPrefix: fmt.Sprintf("emb:%s:%d:", model, dimensions),
		ttl:       ttl,
	}
}

func (e *CachingEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	key := e.keyPrefix + hashText(text)

	if b, ok, err := e.cache.Get(ctx, key); err == nil && ok {
		if vec, derr := decodeVector(b); derr == nil {
			return vec, nil
		}
	}

	vec, err := e.inner.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	if b, derr := encodeVector(vec); derr == nil {
		_ = e.cache.Set(ctx, key, b, e.ttl)
	}
	return vec, nil
}

// EmbedBatch caches each text individually, so a batch that overlaps a previous
// query reuses the cached vectors and only embeds the misses.
func (e *CachingEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		vec, err := e.Embed(ctx, t)
		if err != nil {
			return nil, err
		}
		out[i] = vec
	}
	return out, nil
}

func hashText(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func encodeVector(vec []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, vec); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeVector(b []byte) ([]float32, error) {
	if len(b) == 0 || len(b)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding cache payload length %d", len(b))
	}
	vec := make([]float32, len(b)/4)
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, vec); err != nil {
		return nil, err
	}
	return vec, nil
}
