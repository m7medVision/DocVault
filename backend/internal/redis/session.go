package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SessionData represents session information stored in Redis.
type SessionData struct {
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	DeviceID     string    `json:"device_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// SessionStore manages user sessions in Redis.
type SessionStore struct {
	client *Client
}

// NewSessionStore creates a new SessionStore.
func NewSessionStore(client *Client) *SessionStore {
	return &SessionStore{client: client}
}

// SaveSession stores session data with the specified TTL.
// Supports multiple sessions per user (multi-device).
func (ss *SessionStore) SaveSession(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if data == nil {
		return fmt.Errorf("session data cannot be nil")
	}

	key := fmt.Sprintf("session:%s", sessionID)

	// Serialize session data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Store in Redis with TTL
	err = ss.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Track user's active sessions for multi-device management
	userSessionsKey := fmt.Sprintf("user:sessions:%s", data.UserID)
	err = ss.client.SAdd(ctx, userSessionsKey, sessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to track user session: %w", err)
	}
	
	// Set expiry on user sessions set
	ss.client.Expire(ctx, userSessionsKey, ttl)

	return nil
}

// GetSession retrieves session data by session ID.
func (ss *SessionStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	key := fmt.Sprintf("session:%s", sessionID)

	jsonData, err := ss.client.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// DeleteSession removes a session from Redis.
func (ss *SessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	// Get session to find user ID
	session, err := ss.GetSession(ctx, sessionID)
	if err != nil {
		// Session doesn't exist, that's okay
		return nil
	}

	key := fmt.Sprintf("session:%s", sessionID)
	
	// Delete session data
	err = ss.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user's active sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
	ss.client.SRem(ctx, userSessionsKey, sessionID)

	return nil
}

// GetUserSessions returns all active session IDs for a user.
func (ss *SessionStore) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := fmt.Sprintf("user:sessions:%s", userID)
	
	sessionIDs, err := ss.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessionIDs, nil
}

// DeleteAllUserSessions removes all sessions for a user (logout from all devices).
func (ss *SessionStore) DeleteAllUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	sessionIDs, err := ss.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, sessionID := range sessionIDs {
		if err := ss.DeleteSession(ctx, sessionID); err != nil {
			// Log error but continue deleting other sessions
			continue
		}
	}

	// Clean up user sessions set
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	ss.client.Del(ctx, userSessionsKey)

	return nil
}

// UpdateActivity updates the last activity timestamp for a session.
func (ss *SessionStore) UpdateActivity(ctx context.Context, sessionID string) error {
	session, err := ss.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()

	// Get current TTL to preserve it
	key := fmt.Sprintf("session:%s", sessionID)
	ttl, err := ss.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session TTL: %w", err)
	}

	return ss.SaveSession(ctx, sessionID, session, ttl)
}
