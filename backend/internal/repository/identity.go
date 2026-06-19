package repository

import (
	"context"
	"time"
)

// MemberRecord is a denormalized membership row (membership joined with its
// user) returned by the membership listing queries.
type MemberRecord struct {
	MembershipID string    `json:"membership_id"`
	UserID       string    `json:"user_id"`
	OrgID        string    `json:"org_id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// MembershipRepository provides membership data access.
type MembershipRepository interface {
	ListByOrg(ctx context.Context, tenantID, orgID string) ([]MemberRecord, error)
	GetByID(ctx context.Context, tenantID, membershipID string) (*MemberRecord, error)
	UpdateRole(ctx context.Context, membershipID, role string) error
}

// RegisterWorkspaceParams is the full set of rows created atomically when a new
// user self-registers: a tenant, its default organization, the owner user, and
// the owner membership. IDs are supplied by the caller so the transport can
// build its response without a read-back.
type RegisterWorkspaceParams struct {
	TenantID     string
	TenantName   string
	OrgID        string
	OrgName      string
	UserID       string
	Email        string
	PasswordHash string
	DisplayName  string
	Locale       string
	MembershipID string
	CreatedAt    time.Time
}

// RegistrationRepository owns the transactional creation of a new workspace.
// The unit of work (begin/commit/rollback) lives in the adapter, so the
// transport layer never opens a database transaction itself.
type RegistrationRepository interface {
	RegisterWorkspace(ctx context.Context, params RegisterWorkspaceParams) error
}
