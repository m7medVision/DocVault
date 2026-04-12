package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_NewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		issuer      string
		audience    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid configuration",
			secret:   "this-is-a-very-secure-secret-key-with-32-chars-or-more",
			issuer:   "docvault.app",
			audience: "docvault-api",
			wantErr:  false,
		},
		{
			name:        "empty secret",
			secret:      "",
			issuer:      "docvault.app",
			audience:    "docvault-api",
			wantErr:     true,
			errContains: "secret cannot be empty",
		},
		{
			name:        "short secret",
			secret:      "short",
			issuer:      "docvault.app",
			audience:    "docvault-api",
			wantErr:     true,
			errContains: "must be at least 32 characters",
		},
		{
			name:        "empty issuer",
			secret:      "this-is-a-very-secure-secret-key-with-32-chars-or-more",
			issuer:      "",
			audience:    "docvault-api",
			wantErr:     true,
			errContains: "issuer cannot be empty",
		},
		{
			name:        "empty audience",
			secret:      "this-is-a-very-secure-secret-key-with-32-chars-or-more",
			issuer:      "docvault.app",
			audience:    "",
			wantErr:     true,
			errContains: "audience cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewJWTService(tt.secret, tt.issuer, tt.audience, 15*time.Minute, 720*time.Hour)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, svc)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      string
		email       string
		tenantID    string
		orgID       string
		role        string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid user data",
			userID:   "user-123",
			email:    "test@example.com",
			tenantID: "tenant-456",
			orgID:    "org-789",
			role:     "admin",
			wantErr:  false,
		},
		{
			name:        "missing user ID",
			userID:      "",
			email:       "test@example.com",
			tenantID:    "tenant-456",
			orgID:       "org-789",
			role:        "admin",
			wantErr:     true,
			errContains: "all user fields are required",
		},
		{
			name:        "missing email",
			userID:      "user-123",
			email:       "",
			tenantID:    "tenant-456",
			orgID:       "org-789",
			role:        "admin",
			wantErr:     true,
			errContains: "all user fields are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := svc.GenerateTokenPair(tt.userID, tt.email, tt.tenantID, tt.orgID, tt.role)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, pair)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, pair)
				assert.NotEmpty(t, pair.AccessToken)
				assert.NotEmpty(t, pair.RefreshToken)
				assert.Equal(t, "Bearer", pair.TokenType)
				assert.True(t, pair.ExpiresAt.After(time.Now()))
				assert.True(t, pair.ExpiresAt.Before(time.Now().Add(20*time.Minute)))
			}
		})
	}
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	// Generate a valid token
	pair, err := svc.GenerateTokenPair("user-123", "test@example.com", "tenant-456", "org-789", "admin")
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
		checkClaims func(t *testing.T, claims *TokenClaims)
	}{
		{
			name:    "valid token",
			token:   pair.AccessToken,
			wantErr: false,
			checkClaims: func(t *testing.T, claims *TokenClaims) {
				assert.Equal(t, "user-123", claims.UserID)
				assert.Equal(t, "test@example.com", claims.Email)
				assert.Equal(t, "tenant-456", claims.TenantID)
				assert.Equal(t, "org-789", claims.OrgID)
				assert.Equal(t, "admin", claims.Role)
			},
		},
		{
			name:        "empty token",
			token:       "",
			wantErr:     true,
			errContains: "token cannot be empty",
		},
		{
			name:        "invalid token",
			token:       "invalid.token.here",
			wantErr:     true,
			errContains: "failed to parse token",
		},
		{
			name:        "malformed token",
			token:       "not-a-jwt",
			wantErr:     true,
			errContains: "failed to parse token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ValidateAccessToken(tt.token)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, claims)
				if tt.checkClaims != nil {
					tt.checkClaims(t, claims)
				}
			}
		})
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	// Generate a valid token
	pair, err := svc.GenerateTokenPair("user-123", "test@example.com", "tenant-456", "org-789", "admin")
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, userID, tokenID string)
	}{
		{
			name:    "valid refresh token",
			token:   pair.RefreshToken,
			wantErr: false,
			checkResult: func(t *testing.T, userID, tokenID string) {
				assert.Equal(t, "user-123", userID)
				assert.NotEmpty(t, tokenID)
			},
		},
		{
			name:        "empty token",
			token:       "",
			wantErr:     true,
			errContains: "token cannot be empty",
		},
		{
			name:        "invalid token",
			token:       "invalid.token.here",
			wantErr:     true,
			errContains: "failed to parse refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, tokenID, err := svc.ValidateRefreshToken(tt.token)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, userID)
				assert.Empty(t, tokenID)
			} else {
				require.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, userID, tokenID)
				}
			}
		})
	}
}

func TestJWTService_ExtractTokenID(t *testing.T) {
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	// Generate a valid token
	pair, err := svc.GenerateTokenPair("user-123", "test@example.com", "tenant-456", "org-789", "admin")
	require.NoError(t, err)

	t.Run("extract from valid refresh token", func(t *testing.T) {
		tokenID, expiry, err := svc.ExtractTokenID(pair.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenID)
		assert.True(t, expiry.After(time.Now()))
	})

	t.Run("empty token", func(t *testing.T) {
		tokenID, expiry, err := svc.ExtractTokenID("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
		assert.Empty(t, tokenID)
		assert.True(t, expiry.IsZero())
	})
}

func TestJWTService_TokenExpiry(t *testing.T) {
	// Test with very short TTL
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		1*time.Millisecond, // Very short for testing
		1*time.Millisecond,
	)
	require.NoError(t, err)

	pair, err := svc.GenerateTokenPair("user-123", "test@example.com", "tenant-456", "org-789", "admin")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	_, err = svc.ValidateAccessToken(pair.AccessToken)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTService_WrongAudience(t *testing.T) {
	svc, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"docvault-api",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	// Generate token
	pair, err := svc.GenerateTokenPair("user-123", "test@example.com", "tenant-456", "org-789", "admin")
	require.NoError(t, err)

	// Create service with different audience
	svc2, err := NewJWTService(
		"this-is-a-very-secure-secret-key-with-32-chars-or-more",
		"docvault.app",
		"different-audience",
		15*time.Minute,
		720*time.Hour,
	)
	require.NoError(t, err)

	// Try to validate with wrong audience
	_, err = svc2.ValidateAccessToken(pair.AccessToken)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token audience")
}
