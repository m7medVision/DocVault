// Package middleware provides Role-Based Access Control (RBAC).
package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/docvault/backend/internal/authz"
)

const (
	RoleViewer = authz.RoleViewer
	RoleMember = authz.RoleMember
	RoleAdmin  = authz.RoleAdmin
	RoleOwner  = authz.RoleOwner
)

var RolePermissions = map[string]int{
	authz.RoleViewer: 1,
	authz.RoleMember: 2,
	authz.RoleAdmin:  3,
	authz.RoleOwner:  4,
}

// RBACConfig configures RBAC requirements for a route.
type RBACConfig struct {
	// RequireRole specifies the minimum role required.
	RequireRole string

	// AllowedRoles specifies exact roles that can access the endpoint.
	// If set, RequireRole is ignored.
	AllowedRoles []string

	// RequireTenantMatch ensures the user's tenant matches the resource tenant.
	RequireTenantMatch bool
}

// RequireRole middleware checks if the user has at least the specified role.
func RequireRole(minRole string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())

			if !HasMinRole(userRole, minRole) {
				http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware checks if the user has any of the specified roles.
func RequireAnyRole(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())

			if !slices.Contains(roles, userRole) {
				http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwner restricts access to owners only.
func RequireOwner() Middleware {
	return RequireRole(RoleOwner)
}

// RequireAdmin restricts access to admins and owners.
func RequireAdmin() Middleware {
	return RequireRole(RoleAdmin)
}

// RequireMember restricts access to members, admins, and owners.
func RequireMember() Middleware {
	return RequireRole(RoleMember)
}

// HasMinRole checks if a user role meets the minimum required role.
func HasMinRole(userRole, minRole string) bool {
	userLevel, ok := RolePermissions[userRole]
	if !ok {
		return false
	}

	minLevel, ok := RolePermissions[minRole]
	if !ok {
		return false
	}

	return userLevel >= minLevel
}

// CanWrite checks if the user has write permissions (member or higher).
func CanWrite(role string) bool {
	return HasMinRole(role, RoleMember)
}

// CanDelete checks if the user has delete permissions (admin or higher).
func CanDelete(role string) bool {
	return HasMinRole(role, RoleAdmin)
}

// CanManageMembers checks if the user can manage organization members.
func CanManageMembers(role string) bool {
	return HasMinRole(role, RoleAdmin)
}

// CanManageOrg checks if the user can manage organization settings.
func CanManageOrg(role string) bool {
	return HasMinRole(role, RoleOwner)
}

// WithRole adds a role to the request context (for testing).
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

// ResourceOwner middleware checks if the user owns the resource or has admin access.
// This is useful for operations where only the resource owner or an admin can modify.
func ResourceOwner() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			userRole := GetUserRole(r.Context())

			// Admins and owners can access any resource
			if HasMinRole(userRole, RoleAdmin) {
				next.ServeHTTP(w, r)
				return
			}

			// For non-admins, we need to check resource ownership
			// This is typically done in the handler after fetching the resource
			// Add user ID to context for handler-level ownership checks
			ctx := context.WithValue(r.Context(), "auth_user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantIsolation middleware ensures users can only access resources in their tenant.
func TenantIsolation() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userTenant := GetTenantID(r.Context())

			if userTenant == "" {
				http.Error(w, `{"error":"tenant context missing","code":"FORBIDDEN"}`, http.StatusForbidden)
				return
			}

			// Tenant ID is passed to handlers for filtering queries
			ctx := context.WithValue(r.Context(), "filter_tenant_id", userTenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuditAction is a helper to extract the audit action from context.
type AuditAction string

const (
	ActionView   AuditAction = "view"
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
	ActionShare  AuditAction = "share"
)

// RBACConfig returns RBAC configuration for common API patterns.
func NewRBACConfig(requireRole string, allowedRoles ...string) RBACConfig {
	if len(allowedRoles) > 0 {
		return RBACConfig{
			AllowedRoles:       allowedRoles,
			RequireTenantMatch: true,
		}
	}
	return RBACConfig{
		RequireRole:        requireRole,
		RequireTenantMatch: true,
	}
}
