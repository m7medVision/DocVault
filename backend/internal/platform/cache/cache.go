// Package cache defines the caching port consumed by application services and
// its adapters. The port is deliberately small and the default implementation
// is a no-op, so services can take a Cache dependency and adopt caching
// incrementally without any behavior change until a real adapter is wired in.
package cache

import (
	"context"
	"time"
)

// Cache is a byte-oriented key/value cache. Callers serialize their own values
// (e.g. JSON) so the port stays free of any type or domain coupling. A cache
// miss is reported as (nil, false, nil); only genuine backend failures return a
// non-nil error.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	// DelByPrefix removes every key sharing the given prefix. It is intended
	// for invalidation fan-out (e.g. all of a tenant's cached entries) and is
	// not a hot-path operation.
	DelByPrefix(ctx context.Context, prefix string) error
}

// noop is a Cache that stores nothing and always misses. It lets a service
// depend on Cache while caching is effectively disabled.
type noop struct{}

// NewNoop returns a Cache that never stores or returns anything.
func NewNoop() Cache { return noop{} }

func (noop) Get(context.Context, string) ([]byte, bool, error)        { return nil, false, nil }
func (noop) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (noop) Del(context.Context, ...string) error                     { return nil }
func (noop) DelByPrefix(context.Context, string) error                { return nil }
