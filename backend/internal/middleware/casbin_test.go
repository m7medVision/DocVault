package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeMiddleware(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("..", "authz", "model.conf"))
	require.NoError(t, err)
	require.NoError(t, authz.EnsureTenantRoleAccess(enforcer, "alice", authz.RoleAdmin, "tenant-1"))

	handler := middleware.Authorize(enforcer, authz.ResourceAdminMembers, authz.ActionInvite)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/members/invite", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "alice")
	ctx = context.WithValue(ctx, middleware.TenantIDKey, "tenant-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthorizeMiddlewareDenied(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("..", "authz", "model.conf"))
	require.NoError(t, err)
	require.NoError(t, authz.EnsureTenantRoleAccess(enforcer, "bob", authz.RoleViewer, "tenant-1"))

	handler := middleware.Authorize(enforcer, authz.ResourceAdminMembers, authz.ActionInvite)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/members/invite", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "bob")
	ctx = context.WithValue(ctx, middleware.TenantIDKey, "tenant-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorizeMiddlewareMissingUser(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("..", "authz", "model.conf"))
	require.NoError(t, err)

	handler := middleware.Authorize(enforcer, authz.ResourceDocuments, authz.ActionRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
