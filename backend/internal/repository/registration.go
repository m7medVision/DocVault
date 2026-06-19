package repository

import (
	"context"
	"fmt"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
// The unit of work (begin/commit/rollback) lives here, in the adapter, so the
// transport layer never opens a database transaction itself.
type RegistrationRepository interface {
	RegisterWorkspace(ctx context.Context, params RegisterWorkspaceParams) error
}

type registrationRepository struct {
	db      *pgxpool.Pool
	queries *sqldb.Queries
}

// NewRegistrationRepository creates a RegistrationRepository over the pool.
func NewRegistrationRepository(db *pgxpool.Pool) RegistrationRepository {
	return &registrationRepository{db: db, queries: sqldb.New(db)}
}

// RegisterWorkspace creates the tenant, organization, owner user, and owner
// membership inside a single transaction. Any failure rolls the whole set back,
// so a half-provisioned workspace can never be observed.
func (r *registrationRepository) RegisterWorkspace(ctx context.Context, p RegisterWorkspaceParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin registration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)
	createdAt := pgtype.Timestamptz{Time: p.CreatedAt, Valid: true}
	passwordHash := p.PasswordHash

	if err := q.CreateTenant(ctx, sqldb.CreateTenantParams{
		ID:        p.TenantID,
		Name:      p.TenantName,
		Plan:      "free",
		CreatedAt: createdAt,
	}); err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	if err := q.CreateOrganization(ctx, sqldb.CreateOrganizationParams{
		ID:        p.OrgID,
		TenantID:  p.TenantID,
		Name:      p.OrgName,
		CreatedAt: createdAt,
	}); err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	if err := q.CreateUser(ctx, sqldb.CreateUserParams{
		ID:                  p.UserID,
		TenantID:            p.TenantID,
		Email:               p.Email,
		PasswordHash:        &passwordHash,
		DisplayName:         p.DisplayName,
		Locale:              p.Locale,
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           createdAt,
	}); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := q.CreateMembership(ctx, sqldb.CreateMembershipParams{
		ID:        p.MembershipID,
		UserID:    p.UserID,
		OrgID:     p.OrgID,
		Role:      sqldb.MembershipRoleOwner,
		CreatedAt: createdAt,
	}); err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit registration transaction: %w", err)
	}
	return nil
}
