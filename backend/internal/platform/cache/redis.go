package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCache adapts a go-redis client to the Cache port. It reuses the
// application's existing Redis client and connection pool.
type redisCache struct {
	rdb *redis.Client
}

var _ Cache = (*redisCache)(nil)

// NewRedis returns a Cache backed by the given go-redis client.
func NewRedis(rdb *redis.Client) Cache {
	return &redisCache{rdb: rdb}
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	v, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DelByPrefix scans the keyspace for prefix* and deletes matches in batches.
// This is O(matched keys) and used only for invalidation fan-out, never on a
// request hot path.
func (c *redisCache) DelByPrefix(ctx context.Context, prefix string) error {
	const batchSize = 256
	iter := c.rdb.Scan(ctx, 0, prefix+"*", batchSize).Iterator()

	batch := make([]string, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := c.rdb.Del(ctx, batch...).Err(); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for iter.Next(ctx) {
		batch = append(batch, iter.Val())
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return flush()
}
