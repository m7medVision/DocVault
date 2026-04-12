package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

// RateLimiter manages rate limiting using Redis.
type RateLimiter struct {
	limiter *redis_rate.Limiter
}

// RateLimit defines a rate limit rule.
type RateLimit struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
}

// Predefined rate limits for different endpoints.
var (
	// Auth endpoints
	LoginRateLimit    = RateLimit{Requests: 5, Window: 15 * time.Minute}  // 5 attempts per 15 min
	RegisterRateLimit = RateLimit{Requests: 3, Window: 1 * time.Hour}     // 3 attempts per hour
	RefreshRateLimit  = RateLimit{Requests: 10, Window: 1 * time.Minute}  // 10 per minute

	// General API rate limit
	GeneralAPIRateLimit = RateLimit{Requests: 100, Window: 1 * time.Minute} // 100 per minute
)

// NewRateLimiter creates a new rate limiter using Redis.
func NewRateLimiter(client *Client) *RateLimiter {
	limiter := redis_rate.NewLimiter(client.Client)
	
	return &RateLimiter{
		limiter: limiter,
	}
}

// Allow checks if a request is allowed under the rate limit.
// Returns true if allowed, false if rate limit exceeded.
// Also returns the current limit info (remaining requests, reset time).
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit RateLimit) (allowed bool, remaining int, resetAt time.Time, err error) {
	if key == "" {
		return false, 0, time.Time{}, fmt.Errorf("rate limit key cannot be empty")
	}

	// Use redis_rate library's sliding window algorithm
	result, err := rl.limiter.Allow(ctx, key, redis_rate.Limit{
		Rate:   limit.Requests,
		Period: limit.Window,
		Burst:  limit.Requests,
	})
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to check rate limit: %w", err)
	}

	allowed = result.Allowed > 0
	remaining = result.Remaining
	resetAt = time.Now().Add(result.RetryAfter)

	return allowed, remaining, resetAt, nil
}

// AllowN checks if N requests are allowed under the rate limit.
func (rl *RateLimiter) AllowN(ctx context.Context, key string, limit RateLimit, n int) (allowed bool, remaining int, resetAt time.Time, err error) {
	if key == "" {
		return false, 0, time.Time{}, fmt.Errorf("rate limit key cannot be empty")
	}

	result, err := rl.limiter.AllowN(ctx, key, redis_rate.Limit{
		Rate:   limit.Requests,
		Period: limit.Window,
		Burst:  limit.Requests,
	}, n)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to check rate limit: %w", err)
	}

	allowed = result.Allowed >= n
	remaining = result.Remaining
	resetAt = time.Now().Add(result.RetryAfter)

	return allowed, remaining, resetAt, nil
}

// Reset resets the rate limit counter for a key (for testing).
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	// Redis_rate stores keys with a specific prefix
	// We need to delete the key manually
	return rl.limiter.Reset(ctx, key)
}

// BuildKey creates a rate limit key for a specific endpoint and identifier.
// Format: "ratelimit:{endpoint}:{identifier}"
// Example: "ratelimit:/api/v1/auth/login:user@example.com"
func BuildRateLimitKey(endpoint string, identifier string) string {
	return fmt.Sprintf("ratelimit:%s:%s", endpoint, identifier)
}
