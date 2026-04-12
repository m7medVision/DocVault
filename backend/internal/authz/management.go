package authz

import (
	"fmt"
	"sort"

	"github.com/casbin/casbin/v3"
)

type PolicyRule struct {
	Subject  string `json:"subject"`
	TenantID string `json:"tenant_id"`
	Object   string `json:"object"`
	Action   string `json:"action"`
}

type RoleBinding struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

func isKnownRole(value string) bool {
	return ValidateRole(value) == nil
}

func ValidateRole(role string) error {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return nil
	default:
		return fmt.Errorf("invalid role %q", role)
	}
}

func AddPolicyRule(enforcer *casbin.Enforcer, subject, tenantID, object, action string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is required")
	}
	if subject == "" || tenantID == "" || object == "" || action == "" {
		return fmt.Errorf("subject, tenant ID, object, and action are required")
	}

	if _, err := enforcer.AddPolicy(subject, tenantID, object, action); err != nil {
		return fmt.Errorf("add policy rule: %w", err)
	}

	return nil
}

func RemovePolicyRule(enforcer *casbin.Enforcer, subject, tenantID, object, action string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is required")
	}
	if subject == "" || tenantID == "" || object == "" || action == "" {
		return fmt.Errorf("subject, tenant ID, object, and action are required")
	}

	if _, err := enforcer.RemovePolicy(subject, tenantID, object, action); err != nil {
		return fmt.Errorf("remove policy rule: %w", err)
	}

	return nil
}

func ListPolicyRules(enforcer *casbin.Enforcer, tenantID string) ([]PolicyRule, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	rules, err := enforcer.GetFilteredPolicy(1, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list policy rules: %w", err)
	}

	policies := make([]PolicyRule, 0, len(rules))
	for _, rule := range rules {
		if len(rule) < 4 {
			continue
		}

		policies = append(policies, PolicyRule{
			Subject:  rule[0],
			TenantID: rule[1],
			Object:   rule[2],
			Action:   rule[3],
		})
	}

	sort.Slice(policies, func(i, j int) bool {
		if policies[i].Subject != policies[j].Subject {
			return policies[i].Subject < policies[j].Subject
		}
		if policies[i].Object != policies[j].Object {
			return policies[i].Object < policies[j].Object
		}
		return policies[i].Action < policies[j].Action
	})

	return policies, nil
}

func ReplaceUserRoles(enforcer *casbin.Enforcer, userID, tenantID string, roles []string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	if _, err := enforcer.DeleteRolesForUserInDomain(userID, tenantID); err != nil {
		return fmt.Errorf("clear existing user roles: %w", err)
	}

	for _, role := range roles {
		if err := ValidateRole(role); err != nil {
			return err
		}
		if _, err := enforcer.AddRoleForUserInDomain(userID, role, tenantID); err != nil {
			return fmt.Errorf("assign user role %s: %w", role, err)
		}
	}

	return nil
}

func AssignRoleToUser(enforcer *casbin.Enforcer, userID, role, tenantID string) ([]string, error) {
	if err := ValidateRole(role); err != nil {
		return nil, err
	}
	if err := SeedTenantPolicies(enforcer, tenantID); err != nil {
		return nil, err
	}

	currentRoles := append([]string(nil), enforcer.GetRolesForUserInDomain(userID, tenantID)...)
	if err := ReplaceUserRoles(enforcer, userID, tenantID, []string{role}); err != nil {
		return nil, err
	}

	return currentRoles, nil
}

func GetRolesForUser(enforcer *casbin.Enforcer, userID, tenantID string) ([]string, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	roles := append([]string(nil), enforcer.GetRolesForUserInDomain(userID, tenantID)...)
	sort.Strings(roles)
	return roles, nil
}

func GetUsersForRole(enforcer *casbin.Enforcer, role, tenantID string) ([]string, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer is required")
	}
	if err := ValidateRole(role); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	users := append([]string(nil), enforcer.GetUsersForRoleInDomain(role, tenantID)...)
	filtered := users[:0]
	for _, user := range users {
		if isKnownRole(user) {
			continue
		}
		filtered = append(filtered, user)
	}
	sort.Strings(filtered)
	return filtered, nil
}

func ListRoleBindings(enforcer *casbin.Enforcer, tenantID string) ([]RoleBinding, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	rules, err := enforcer.GetFilteredGroupingPolicy(2, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list role bindings: %w", err)
	}

	bindings := make([]RoleBinding, 0, len(rules))
	for _, rule := range rules {
		if len(rule) < 3 {
			continue
		}
		if isKnownRole(rule[0]) && isKnownRole(rule[1]) {
			continue
		}

		bindings = append(bindings, RoleBinding{
			UserID:   rule[0],
			Role:     rule[1],
			TenantID: rule[2],
		})
	}

	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Role != bindings[j].Role {
			return bindings[i].Role < bindings[j].Role
		}
		return bindings[i].UserID < bindings[j].UserID
	})

	return bindings, nil
}

func GetPermissionsForUser(enforcer *casbin.Enforcer, userID, tenantID string) ([]Permission, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	rules, err := enforcer.GetImplicitPermissionsForUser(userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get permissions for user: %w", err)
	}

	permissions := make([]Permission, 0, len(rules))
	for _, rule := range rules {
		if len(rule) < 4 {
			continue
		}

		permissions = append(permissions, Permission{
			Role:   rule[0],
			Object: rule[2],
			Action: rule[3],
		})
	}

	sort.Slice(permissions, func(i, j int) bool {
		if permissions[i].Role != permissions[j].Role {
			return permissions[i].Role < permissions[j].Role
		}
		if permissions[i].Object != permissions[j].Object {
			return permissions[i].Object < permissions[j].Object
		}
		return permissions[i].Action < permissions[j].Action
	})

	return permissions, nil
}
