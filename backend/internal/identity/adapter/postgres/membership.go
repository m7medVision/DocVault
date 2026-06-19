package postgres

import (
	"context"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/identity"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MembershipRepository handles membership data access. It satisfies the
// repository.MembershipRepository contract; the composition root binds this
// concrete type to that interface.
type MembershipRepository struct {
	queries sqldb.Querier
}

// NewMembershipRepository creates a postgres-backed membership repository.
func NewMembershipRepository(db *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{queries: sqldb.New(db)}
}

func (r *MembershipRepository) ListByOrg(ctx context.Context, tenantID, orgID string) ([]repository.MemberRecord, error) {
	var records []repository.MemberRecord
	if orgID != "" {
		rows, err := r.queries.ListMembershipsByOrg(ctx, sqldb.ListMembershipsByOrgParams{
			TenantID: tenantID,
			OrgID:    orgID,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			records = append(records, repository.MemberRecord{
				MembershipID: row.MembershipID,
				UserID:       row.UserID,
				OrgID:        row.OrgID,
				Email:        row.Email,
				DisplayName:  row.DisplayName,
				Role:         string(row.Role),
				CreatedAt:    row.CreatedAt.Time,
			})
		}
		return records, nil
	}

	rows, err := r.queries.ListMembershipsByTenant(ctx, sqldb.ListMembershipsByTenantParams{TenantID: tenantID})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		records = append(records, repository.MemberRecord{
			MembershipID: row.MembershipID,
			UserID:       row.UserID,
			OrgID:        row.OrgID,
			Email:        row.Email,
			DisplayName:  row.DisplayName,
			Role:         string(row.Role),
			CreatedAt:    row.CreatedAt.Time,
		})
	}
	return records, nil
}

func (r *MembershipRepository) GetByID(ctx context.Context, tenantID, membershipID string) (*repository.MemberRecord, error) {
	row, err := r.queries.GetMembershipByID(ctx, sqldb.GetMembershipByIDParams{
		TenantID: tenantID,
		ID:       membershipID,
	})
	if err != nil {
		return nil, err
	}

	record := repository.MemberRecord{
		MembershipID: row.MembershipID,
		UserID:       row.UserID,
		OrgID:        row.OrgID,
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		Role:         string(row.Role),
		CreatedAt:    row.CreatedAt.Time,
	}
	return &record, nil
}

func (r *MembershipRepository) UpdateRole(ctx context.Context, membershipID, role string) error {
	rowsAffected, err := r.queries.UpdateMembershipRole(ctx, sqldb.UpdateMembershipRoleParams{
		Role: sqldb.MembershipRole(role),
		ID:   membershipID,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MembershipRepository) Create(ctx context.Context, m *model.Membership) error {
	return r.queries.CreateMembership(ctx, sqldb.CreateMembershipParams{
		ID:        m.ID,
		UserID:    m.UserID,
		OrgID:     m.OrgID,
		Role:      sqldb.MembershipRole(m.Role),
		CreatedAt: pgtype.Timestamptz{Time: m.CreatedAt, Valid: true},
	})
}
