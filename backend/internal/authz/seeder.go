package authz

import (
	"fmt"

	"github.com/casbin/casbin/v3"
)

func SeedTenantPolicies(enforcer *casbin.Enforcer, tenantID string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	for _, link := range defaultHierarchy {
		if _, err := enforcer.AddGroupingPolicy(link[0], link[1], tenantID); err != nil {
			return fmt.Errorf("seed role hierarchy %s -> %s: %w", link[0], link[1], err)
		}
	}

	for _, permission := range defaultPermissions {
		if _, err := enforcer.AddPolicy(permission.Role, tenantID, permission.Object, permission.Action); err != nil {
			return fmt.Errorf("seed permission %s/%s/%s: %w", permission.Role, permission.Object, permission.Action, err)
		}
	}

	return nil
}

func EnsureUserRoleBinding(enforcer *casbin.Enforcer, userID, role, tenantID string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if role == "" {
		return fmt.Errorf("role is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if err := ValidateRole(role); err != nil {
		return err
	}

	if err := ReplaceUserRoles(enforcer, userID, tenantID, []string{role}); err != nil {
		return fmt.Errorf("bind user role in domain: %w", err)
	}

	return nil
}

func EnsureTenantRoleAccess(enforcer *casbin.Enforcer, userID, role, tenantID string) error {
	if err := SeedTenantPolicies(enforcer, tenantID); err != nil {
		return err
	}

	if err := EnsureUserRoleBinding(enforcer, userID, role, tenantID); err != nil {
		return err
	}

	if enforcer.GetAdapter() != nil {
		if err := enforcer.SavePolicy(); err != nil {
			return fmt.Errorf("save tenant policies: %w", err)
		}
	}

	return nil
}
