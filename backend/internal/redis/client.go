// Package redis provides Redis client and connection management for DocVault.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality.
type Client struct {
	*redis.Client
	config *Config
}

// Config holds Redis connection configuration.
type Config struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewClient creates a new Redis client with the provided configuration.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		// Connection pool settings
		MinIdleConns: 5,
		MaxIdleConns: 10,

		// Retry configuration
		MaxRetryBackoff: 512 * time.Millisecond,
		MinRetryBackoff: 8 * time.Millisecond,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}

	slog.Info("redis client initialized", "addr", addr, "db", cfg.DB, "pool_size", cfg.PoolSize)

	return &Client{
		Client: rdb,
		config: cfg,
	}, nil
}

// HealthCheck verifies Redis connectivity.
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

// Close gracefully closes the Redis client connection.
func (c *Client) Close() error {
	slog.Info("closing redis client connection")
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}
	return nil
}

// Stats returns current connection pool statistics.
func (c *Client) Stats() *redis.PoolStats {
	return c.PoolStats()
}
