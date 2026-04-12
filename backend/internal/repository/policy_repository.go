package repository

import (
	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/authz"
)

type AuthorizerProvider interface {
	GetAuthorizer() authz.Authorizer
}

type PolicyRepository interface {
	ListPolicies(tenantID string) ([]authz.PolicyRule, error)
	AddPolicy(subject, tenantID, object, action string) error
	RemovePolicy(subject, tenantID, object, action string) error
}

type policyRepository struct {
	enforcer *casbin.Enforcer
}

func NewPolicyRepository(enforcer *casbin.Enforcer) PolicyRepository {
	return &policyRepository{enforcer: enforcer}
}

func (r *policyRepository) ListPolicies(tenantID string) ([]authz.PolicyRule, error) {
	return authz.ListPolicyRules(r.enforcer, tenantID)
}

func (r *policyRepository) AddPolicy(subject, tenantID, object, action string) error {
	return authz.AddPolicyRule(r.enforcer, subject, tenantID, object, action)
}

func (r *policyRepository) RemovePolicy(subject, tenantID, object, action string) error {
	return authz.RemovePolicyRule(r.enforcer, subject, tenantID, object, action)
}
