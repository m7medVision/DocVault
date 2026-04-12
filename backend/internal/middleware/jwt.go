// Package middleware provides HTTP middleware for authentication and authorization.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/redis"
)

const (
	jwtClaimsKey contextKey = "jwt_claims"
)

// JWTMiddleware validates JWT tokens and adds user claims to the request context.
type JWTMiddleware struct {
	jwtService      *auth.JWTService
	tokenBlacklist  *redis.TokenBlacklist
	publicEndpoints []string // Endpoints that don't require authentication
}

// NewJWTMiddleware creates a new JWT middleware.
func NewJWTMiddleware(jwtService *auth.JWTService, tokenBlacklist *redis.TokenBlacklist) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService:     jwtService,
		tokenBlacklist: tokenBlacklist,
		publicEndpoints: []string{
			"/health",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/auth/logout",
			"/metrics", // Prometheus metrics (optional, can be restricted)
		},
	}
}

// Authenticate is the HTTP middleware handler.
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a public endpoint
		if m.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Check for Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		ctx := r.Context()

		// Validate token and extract claims
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Add claims to request context
		ctx = context.WithValue(ctx, jwtClaimsKey, claims)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, OrgIDKey, claims.OrgID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isPublicEndpoint checks if the path is a public endpoint.
func (m *JWTMiddleware) isPublicEndpoint(path string) bool {
	for _, endpoint := range m.publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

// GetUserClaims extracts user claims from the request context.
// Returns nil if no claims are present.
func GetUserClaims(r *http.Request) *auth.TokenClaims {
	claims, ok := r.Context().Value(jwtClaimsKey).(*auth.TokenClaims)
	if !ok {
		return nil
	}
	return claims
}

// RequireAuth is a helper to get user claims or return 401.
// Use this in handlers that require authentication.
func RequireAuth(r *http.Request) (*auth.TokenClaims, error) {
	claims := GetUserClaims(r)
	if claims == nil {
		return nil, http.ErrNoCookie // Reusing error for "not authenticated"
	}
	return claims, nil
}

// respondError sends a JSON error response.
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Simple JSON response without dependencies
	w.Write([]byte(`{"error":"` + message + `"}`))
}
