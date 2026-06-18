package authz_test

import (
	"path/filepath"
	"testing"

	"github.com/docvault/backend/internal/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedTenantPolicies(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	err = authz.SeedTenantPolicies(enforcer, "tenant-1")
	require.NoError(t, err)

	_, err = enforcer.AddRoleForUserInDomain("alice", authz.RoleOwner, "tenant-1")
	require.NoError(t, err)
	_, err = enforcer.AddRoleForUserInDomain("bob", authz.RoleMember, "tenant-1")
	require.NoError(t, err)
	_, err = enforcer.AddRoleForUserInDomain("carol", authz.RoleViewer, "tenant-1")
	require.NoError(t, err)

	tests := []struct {
		name     string
		subject  string
		domain   string
		object   string
		action   string
		expected bool
	}{
		{"owner inherits viewer read", "alice", "tenant-1", authz.ResourceDocuments, authz.ActionRead, true},
		{"owner can delete documents", "alice", "tenant-1", authz.ResourceDocuments, authz.ActionDelete, true},
		{"owner can update member role", "alice", "tenant-1", authz.ResourceAdminMembers, authz.ActionUpdateRole, true},
		{"member can write documents", "bob", "tenant-1", authz.ResourceDocuments, authz.ActionWrite, true},
		{"member cannot delete documents", "bob", "tenant-1", authz.ResourceDocuments, authz.ActionDelete, false},
		{"viewer can read search", "carol", "tenant-1", authz.ResourceSearch, authz.ActionRead, true},
		{"viewer cannot invite members", "carol", "tenant-1", authz.ResourceAdminMembers, authz.ActionInvite, false},
		{"cross-tenant access denied", "alice", "tenant-2", authz.ResourceDocuments, authz.ActionRead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.domain, tt.object, tt.action)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

func TestSeedTenantPoliciesIsIdempotent(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	require.NoError(t, authz.SeedTenantPolicies(enforcer, "tenant-1"))
	require.NoError(t, authz.SeedTenantPolicies(enforcer, "tenant-1"))

	policies, err := enforcer.GetPolicy()
	require.NoError(t, err)
	assert.Len(t, policies, 24)

	groupingPolicies, err := enforcer.GetGroupingPolicy()
	require.NoError(t, err)
	assert.Len(t, groupingPolicies, 3)
}

func TestSeedTenantPoliciesValidation(t *testing.T) {
	_, err := authz.NewMemoryEnforcer("")
	require.Error(t, err)

	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	err = authz.SeedTenantPolicies(nil, "tenant-1")
	require.Error(t, err)

	err = authz.SeedTenantPolicies(enforcer, "")
	require.Error(t, err)
}

func TestEnsureTenantRoleAccess(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	require.NoError(t, authz.EnsureTenantRoleAccess(enforcer, "alice", authz.RoleAdmin, "tenant-1"))

	allowed, err := enforcer.Enforce("alice", "tenant-1", authz.ResourceAdminMembers, authz.ActionInvite)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestEnsureTenantRoleAccessReplacesPreviousRole(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	require.NoError(t, authz.EnsureTenantRoleAccess(enforcer, "alice", authz.RoleAdmin, "tenant-1"))
	require.NoError(t, authz.EnsureTenantRoleAccess(enforcer, "alice", authz.RoleViewer, "tenant-1"))

	roles, err := authz.GetRolesForUser(enforcer, "alice", "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, []string{authz.RoleViewer}, roles)

	allowed, err := enforcer.Enforce("alice", "tenant-1", authz.ResourceAdminMembers, authz.ActionInvite)
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestManagementHelpers(t *testing.T) {
	enforcer, err := authz.NewMemoryEnforcer(filepath.Join("model.conf"))
	require.NoError(t, err)

	require.NoError(t, authz.SeedTenantPolicies(enforcer, "tenant-1"))
	require.NoError(t, authz.AddPolicyRule(enforcer, authz.RoleAdmin, "tenant-1", authz.ResourceAdminPolicies, authz.ActionRead))
	_, err = authz.AssignRoleToUser(enforcer, "alice", authz.RoleAdmin, "tenant-1")
	require.NoError(t, err)

	policies, err := authz.ListPolicyRules(enforcer, "tenant-1")
	require.NoError(t, err)
	assert.NotEmpty(t, policies)

	bindings, err := authz.ListRoleBindings(enforcer, "tenant-1")
	require.NoError(t, err)
	assert.Contains(t, bindings, authz.RoleBinding{UserID: "alice", Role: authz.RoleAdmin, TenantID: "tenant-1"})

	permissions, err := authz.GetPermissionsForUser(enforcer, "alice", "tenant-1")
	require.NoError(t, err)
	assert.NotEmpty(t, permissions)

	users, err := authz.GetUsersForRole(enforcer, authz.RoleAdmin, "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, []string{"alice"}, users)

	require.NoError(t, authz.RemovePolicyRule(enforcer, authz.RoleAdmin, "tenant-1", authz.ResourceAdminPolicies, authz.ActionRead))
	policies, err = authz.ListPolicyRules(enforcer, "tenant-1")
	require.NoError(t, err)
	for _, policy := range policies {
		assert.False(t, policy.Subject == authz.RoleAdmin && policy.Object == authz.ResourceAdminPolicies && policy.Action == authz.ActionRead)
	}
}
