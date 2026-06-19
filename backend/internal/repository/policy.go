package repository

import "github.com/docvault/backend/internal/authz"

// PolicyRepository provides access to the Casbin-backed RBAC policy store,
// scoped per tenant. The concrete implementation lives in the identity context's
// casbin adapter; this contract is what the transport layer depends on.
type PolicyRepository interface {
	ListPolicies(tenantID string) ([]authz.PolicyRule, error)
	AddPolicy(subject, tenantID, object, action string) error
	RemovePolicy(subject, tenantID, object, action string) error
}
