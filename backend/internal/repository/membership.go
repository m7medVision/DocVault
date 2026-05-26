package repository

import (
	"context"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberRecord struct {
	MembershipID string    `json:"membership_id"`
	UserID       string    `json:"user_id"`
	OrgID        string    `json:"org_id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type MembershipRepository interface {
	ListByOrg(ctx context.Context, tenantID, orgID string) ([]MemberRecord, error)
	GetByID(ctx context.Context, tenantID, membershipID string) (*MemberRecord, error)
	UpdateRole(ctx context.Context, membershipID, role string) error
}

type membershipRepository struct {
	queries sqldb.Querier
}

func NewMembershipRepository(db *pgxpool.Pool) MembershipRepository {
	return &membershipRepository{queries: sqldb.New(db)}
}

func (r *membershipRepository) ListByOrg(ctx context.Context, tenantID, orgID string) ([]MemberRecord, error) {
	var records []MemberRecord
	if orgID != "" {
		rows, err := r.queries.ListMembershipsByOrg(ctx, sqldb.ListMembershipsByOrgParams{
			TenantID: tenantID,
			OrgID:    orgID,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			records = append(records, MemberRecord{
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
		records = append(records, MemberRecord{
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

func (r *membershipRepository) GetByID(ctx context.Context, tenantID, membershipID string) (*MemberRecord, error) {
	row, err := r.queries.GetMembershipByID(ctx, sqldb.GetMembershipByIDParams{
		TenantID: tenantID,
		ID:       membershipID,
	})
	if err != nil {
		return nil, err
	}

	record := MemberRecord{
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

func (r *membershipRepository) UpdateRole(ctx context.Context, membershipID, role string) error {
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

func (r *membershipRepository) Create(ctx context.Context, m *model.Membership) error {
	return r.queries.CreateMembership(ctx, sqldb.CreateMembershipParams{
		ID:        m.ID,
		UserID:    m.UserID,
		OrgID:     m.OrgID,
		Role:      sqldb.MembershipRole(m.Role),
		CreatedAt: pgtype.Timestamptz{Time: m.CreatedAt, Valid: true},
	})
}
