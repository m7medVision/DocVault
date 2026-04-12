package redis

import (
	"context"
	"testing"
	"time"
)

func TestClientConnection(t *testing.T) {
	cfg := &Config{
		Host:         "localhost",
		Port:         "6379",
		Password:     "changeme",
		DB:           1, // Use DB 1 for tests
		PoolSize:     5,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test PING
	if err := client.HealthCheck(ctx); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// Test SET/GET
	key := "test:key"
	value := "test_value"
	
	err = client.Set(ctx, key, value, 10*time.Second).Err()
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	got, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if got != value {
		t.Errorf("Expected %s, got %s", value, got)
	}

	// Cleanup
	client.Del(ctx, key)
}

func TestTokenBlacklist(t *testing.T) {
	cfg := &Config{
		Host:         "localhost",
		Port:         "6379",
		Password:     "changeme",
		DB:           1,
		PoolSize:     5,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	blacklist := NewTokenBlacklist(client)
	ctx := context.Background()

	tokenID := "test-token-123"

	// Test blacklist
	err = blacklist.BlacklistToken(ctx, tokenID, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to blacklist token: %v", err)
	}

	// Check if blacklisted
	isBlacklisted, err := blacklist.IsBlacklisted(ctx, tokenID)
	if err != nil {
		t.Fatalf("Failed to check blacklist: %v", err)
	}

	if !isBlacklisted {
		t.Error("Token should be blacklisted")
	}

	// Test non-blacklisted token
	isBlacklisted, err = blacklist.IsBlacklisted(ctx, "non-existent-token")
	if err != nil {
		t.Fatalf("Failed to check blacklist: %v", err)
	}

	if isBlacklisted {
		t.Error("Non-existent token should not be blacklisted")
	}

	// Cleanup
	blacklist.RemoveToken(ctx, tokenID)
}

func TestSessionStore(t *testing.T) {
	cfg := &Config{
		Host:         "localhost",
		Port:         "6379",
		Password:     "changeme",
		DB:           1,
		PoolSize:     5,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	store := NewSessionStore(client)
	ctx := context.Background()

	sessionID := "test-session-456"
	sessionData := &SessionData{
		UserID:       "user-123",
		TenantID:     "tenant-456",
		Email:        "test@example.com",
		Role:         "admin",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Save session
	err = store.SaveSession(ctx, sessionID, sessionData, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Get session
	retrieved, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.UserID != sessionData.UserID {
		t.Errorf("Expected UserID %s, got %s", sessionData.UserID, retrieved.UserID)
	}

	// Delete session
	err = store.DeleteSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deletion
	_, err = store.GetSession(ctx, sessionID)
	if err == nil {
		t.Error("Session should have been deleted")
	}
}

func TestRateLimiter(t *testing.T) {
	cfg := &Config{
		Host:         "localhost",
		Port:         "6379",
		Password:     "changeme",
		DB:           1,
		PoolSize:     5,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	limiter := NewRateLimiter(client)
	ctx := context.Background()

	key := "test:ratelimit:user123"
	limit := RateLimit{Requests: 3, Window: 10 * time.Second}

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, _, _, err := limiter.Allow(ctx, key, limit)
		if err != nil {
			t.Fatalf("Rate limit check failed: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	allowed, _, _, err := limiter.Allow(ctx, key, limit)
	if err != nil {
		t.Fatalf("Rate limit check failed: %v", err)
	}
	if allowed {
		t.Error("Request 4 should be denied (rate limit exceeded)")
	}

	// Cleanup
	limiter.Reset(ctx, key)
}
