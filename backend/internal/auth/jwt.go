// Package auth provides JWT token generation and validation for user authentication.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the JWT claims structure for DocVault.
type TokenClaims struct {
	UserID   string `json:"sub"`          // Subject (user ID)
	Email    string `json:"email"`        // User email
	TenantID string `json:"tenant_id"`    // Tenant ID for isolation
	OrgID    string `json:"org_id"`       // Organization ID (primary org)
	Role     string `json:"role"`         // User role (owner|admin|member|viewer)
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"` // Always "Bearer"
}

// JWTService handles JWT token operations.
type JWTService struct {
	secret         []byte
	issuer         string
	audience       string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService creates a new JWT service.
func NewJWTService(secret, issuer, audience string, accessTTL, refreshTTL time.Duration) (*JWTService, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 characters for security")
	}
	if issuer == "" {
		return nil, fmt.Errorf("JWT issuer cannot be empty")
	}
	if audience == "" {
		return nil, fmt.Errorf("JWT audience cannot be empty")
	}

	return &JWTService{
		secret:          []byte(secret),
		issuer:          issuer,
		audience:        audience,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}, nil
}

// GenerateTokenPair creates both access and refresh tokens for a user.
func (s *JWTService) GenerateTokenPair(userID, email, tenantID, orgID, role string) (*TokenPair, error) {
	if userID == "" || email == "" || tenantID == "" || orgID == "" || role == "" {
		return nil, fmt.Errorf("all user fields are required for token generation")
	}

	now := time.Now()
	accessExpiry := now.Add(s.accessTokenTTL)

	// Generate access token
	accessClaims := TokenClaims{
		UserID:   userID,
		Email:    email,
		TenantID: tenantID,
		OrgID:    orgID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (longer TTL, minimal claims)
	refreshExpiry := now.Add(s.refreshTokenTTL)
	refreshClaims := jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Audience:  jwt.ClaimStrings{s.audience},
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        generateTokenID(), // Unique ID for blacklisting
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *JWTService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify audience (check if our audience is in the token's audience list)
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == s.audience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, fmt.Errorf("invalid token audience")
	}

	// Verify issuer
	if claims.Issuer != s.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID and token ID.
func (s *JWTService) ValidateRefreshToken(tokenString string) (userID, tokenID string, err error) {
	if tokenString == "" {
		return "", "", fmt.Errorf("token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return "", "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token claims")
	}

	// Verify audience (check if our audience is in the token's audience list)
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == s.audience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return "", "", fmt.Errorf("invalid token audience")
	}

	// Verify issuer
	if claims.Issuer != s.issuer {
		return "", "", fmt.Errorf("invalid token issuer")
	}

	if claims.Subject == "" {
		return "", "", fmt.Errorf("refresh token missing subject")
	}

	if claims.ID == "" {
		return "", "", fmt.Errorf("refresh token missing ID")
	}

	return claims.Subject, claims.ID, nil
}

// ExtractTokenID extracts the JWT ID (jti) from a refresh token without full validation.
// Used for blacklisting tokens during logout.
func (s *JWTService) ExtractTokenID(tokenString string) (string, time.Time, error) {
	if tokenString == "" {
		return "", time.Time{}, fmt.Errorf("token cannot be empty")
	}

	// Parse without validation (we just need the ID and expiry for blacklisting)
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", time.Time{}, fmt.Errorf("invalid token claims")
	}

	if claims.ID == "" {
		return "", time.Time{}, fmt.Errorf("token missing ID")
	}

	if claims.ExpiresAt == nil {
		return "", time.Time{}, fmt.Errorf("token missing expiration")
	}

	return claims.ID, claims.ExpiresAt.Time, nil
}

// generateTokenID creates a cryptographically secure random ID for refresh tokens.
func generateTokenID() string {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails (should never happen)
		return fmt.Sprintf("token_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}
