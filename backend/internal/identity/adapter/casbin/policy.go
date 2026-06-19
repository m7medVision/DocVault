// Package casbin is the identity bounded context's RBAC policy adapter. It wraps
// the Casbin enforcer behind the repository.PolicyRepository contract.
package casbin

import (
	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/authz"
)

// PolicyRepository is the Casbin-backed RBAC policy store. It satisfies the
// repository.PolicyRepository contract; the composition root binds this concrete
// type to that interface.
type PolicyRepository struct {
	enforcer *casbin.Enforcer
}

// NewPolicyRepository creates a policy repository over the Casbin enforcer.
func NewPolicyRepository(enforcer *casbin.Enforcer) *PolicyRepository {
	return &PolicyRepository{enforcer: enforcer}
}

func (r *PolicyRepository) ListPolicies(tenantID string) ([]authz.PolicyRule, error) {
	return authz.ListPolicyRules(r.enforcer, tenantID)
}

func (r *PolicyRepository) AddPolicy(subject, tenantID, object, action string) error {
	return authz.AddPolicyRule(r.enforcer, subject, tenantID, object, action)
}

func (r *PolicyRepository) RemovePolicy(subject, tenantID, object, action string) error {
	return authz.RemovePolicyRule(r.enforcer, subject, tenantID, object, action)
}
