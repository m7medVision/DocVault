package postgres

import (
	"context"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegistrationRepository handles transactional workspace registration. It
// satisfies the repository.RegistrationRepository contract; the composition
// root binds this concrete type to that interface.
type RegistrationRepository struct {
	db      *pgxpool.Pool
	queries *sqldb.Queries
}

// NewRegistrationRepository creates a postgres-backed registration repository.
func NewRegistrationRepository(db *pgxpool.Pool) *RegistrationRepository {
	return &RegistrationRepository{db: db, queries: sqldb.New(db)}
}

// RegisterWorkspace creates the tenant, organization, owner user, and owner
// membership inside a single transaction. Any failure rolls the whole set back,
// so a half-provisioned workspace can never be observed.
func (r *RegistrationRepository) RegisterWorkspace(ctx context.Context, p repository.RegisterWorkspaceParams) error {
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
