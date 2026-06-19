package cache

import (
	"context"
	"testing"
	"time"
)

func TestNoopAlwaysMisses(t *testing.T) {
	ctx := context.Background()
	c := NewNoop()

	if err := c.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	val, ok, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if ok {
		t.Fatal("noop cache must always report a miss")
	}
	if val != nil {
		t.Fatalf("noop cache must return nil value, got %q", val)
	}

	if err := c.Del(ctx, "k"); err != nil {
		t.Fatalf("Del returned error: %v", err)
	}
	if err := c.DelByPrefix(ctx, "k"); err != nil {
		t.Fatalf("DelByPrefix returned error: %v", err)
	}
}
