package redis

import (
	"context"
	"fmt"
	"time"
)

// TokenBlacklist manages blacklisted JWT tokens for logout functionality.
type TokenBlacklist struct {
	client *Client
}

// NewTokenBlacklist creates a new TokenBlacklist service.
func NewTokenBlacklist(client *Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// BlacklistToken adds a token to the blacklist with the specified TTL.
// The TTL should match the token's expiration time.
func (tb *TokenBlacklist) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if tokenID == "" {
		return fmt.Errorf("token ID cannot be empty")
	}

	key := fmt.Sprintf("blacklist:token:%s", tokenID)

	// Store token in Redis with expiration matching token TTL
	err := tb.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is in the blacklist.
func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, fmt.Errorf("token ID cannot be empty")
	}

	key := fmt.Sprintf("blacklist:token:%s", tokenID)

	exists, err := tb.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists > 0, nil
}

// RemoveToken removes a token from the blacklist (for testing purposes).
func (tb *TokenBlacklist) RemoveToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)

	err := tb.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}
