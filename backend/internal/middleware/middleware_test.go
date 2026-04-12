package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docvault/backend/internal/middleware"
)

// TestHasMinRole tests the role hierarchy check.
func TestHasMinRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		minRole  string
		expected bool
	}{
		{"viewer can access viewer", "viewer", "viewer", true},
		{"viewer cannot access member", "viewer", "member", false},
		{"viewer cannot access admin", "viewer", "admin", false},
		{"viewer cannot access owner", "viewer", "owner", false},
		{"member can access viewer", "member", "viewer", true},
		{"member can access member", "member", "member", true},
		{"member cannot access admin", "member", "admin", false},
		{"member cannot access owner", "member", "owner", false},
		{"admin can access viewer", "admin", "viewer", true},
		{"admin can access member", "admin", "member", true},
		{"admin can access admin", "admin", "admin", true},
		{"admin cannot access owner", "admin", "owner", false},
		{"owner can access viewer", "owner", "viewer", true},
		{"owner can access member", "owner", "member", true},
		{"owner can access admin", "owner", "admin", true},
		{"owner can access owner", "owner", "owner", true},
		{"empty role cannot access anything", "", "viewer", false},
		{"invalid role cannot access anything", "invalid", "viewer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.HasMinRole(tt.userRole, tt.minRole)
			if result != tt.expected {
				t.Errorf("HasMinRole(%q, %q) = %v, want %v", tt.userRole, tt.minRole, result, tt.expected)
			}
		})
	}
}

// TestCanWrite tests the write permission check.
func TestCanWrite(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"viewer", false},
		{"member", true},
		{"admin", true},
		{"owner", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			result := middleware.CanWrite(tt.role)
			if result != tt.expected {
				t.Errorf("CanWrite(%q) = %v, want %v", tt.role, result, tt.expected)
			}
		})
	}
}

// TestCanDelete tests the delete permission check.
func TestCanDelete(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"viewer", false},
		{"member", false},
		{"admin", true},
		{"owner", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			result := middleware.CanDelete(tt.role)
			if result != tt.expected {
				t.Errorf("CanDelete(%q) = %v, want %v", tt.role, result, tt.expected)
			}
		})
	}
}

// TestWithRole tests adding role to context.
func TestWithRole(t *testing.T) {
	ctx := context.Background()
	role := "admin"

	newCtx := middleware.WithRole(ctx, role)

	if middleware.GetUserRole(newCtx) != role {
		t.Errorf("GetUserRole(newCtx) = %q, want %q", middleware.GetUserRole(newCtx), role)
	}
}

// TestContextGetters tests the context getter functions with empty context.
func TestContextGetters(t *testing.T) {
	ctx := context.Background()

	if middleware.GetUserID(ctx) != "" {
		t.Error("GetUserID on empty context should return empty string")
	}

	if middleware.GetTenantID(ctx) != "" {
		t.Error("GetTenantID on empty context should return empty string")
	}

	if middleware.GetOrgID(ctx) != "" {
		t.Error("GetOrgID on empty context should return empty string")
	}

	if middleware.GetUserRole(ctx) != "" {
		t.Error("GetUserRole on empty context should return empty string")
	}
}

// TestCORSMiddleware tests the CORS middleware.
func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := middleware.CORS([]string{"http://localhost:3000"})(handler)

	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("missing Access-Control-Allow-Origin header")
	}

	// Test regular request with allowed origin
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w = httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Test request with disallowed origin
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w = httptest.NewRecorder()

	corsHandler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Error("should not set Access-Control-Allow-Origin for disallowed origin")
	}
}
