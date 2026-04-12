package repository

import (
	"context"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type memberRecord struct {
	MembershipID string
	UserID       string
	OrgID        string
	Email        string
	DisplayName  string
	Role         string
	CreatedAt    time.Time
}

func (r *memberRecord) toModel() MemberRecord {
	return MemberRecord{
		MembershipID: r.MembershipID,
		UserID:       r.UserID,
		OrgID:        r.OrgID,
		Email:        r.Email,
		DisplayName:  r.DisplayName,
		Role:         r.Role,
		CreatedAt:    r.CreatedAt,
	}
}

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
	db *pgxpool.Pool
}

func NewMembershipRepository(db *pgxpool.Pool) MembershipRepository {
	return &membershipRepository{db: db}
}

func (r *membershipRepository) ListByOrg(ctx context.Context, tenantID, orgID string) ([]MemberRecord, error) {
	query := `
		SELECT m.id, m.user_id, m.org_id, u.email, u.display_name, m.role, m.created_at
		FROM memberships m
		JOIN users u ON u.id = m.user_id
		JOIN organizations o ON o.id = m.org_id
		WHERE o.tenant_id = $1`
	args := []interface{}{tenantID}

	if orgID != "" {
		query += ` AND m.org_id = $2`
		args = append(args, orgID)
	}

	query += ` ORDER BY m.created_at ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []MemberRecord
	for rows.Next() {
		var mr memberRecord
		if err := rows.Scan(
			&mr.MembershipID,
			&mr.UserID,
			&mr.OrgID,
			&mr.Email,
			&mr.DisplayName,
			&mr.Role,
			&mr.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, mr.toModel())
	}

	return records, rows.Err()
}

func (r *membershipRepository) GetByID(ctx context.Context, tenantID, membershipID string) (*MemberRecord, error) {
	var mr memberRecord
	err := r.db.QueryRow(ctx, `
		SELECT m.id, m.user_id, m.org_id, u.email, u.display_name, m.role, m.created_at
		FROM memberships m
		JOIN users u ON u.id = m.user_id
		JOIN organizations o ON o.id = m.org_id
		WHERE o.tenant_id = $1 AND m.id = $2`, tenantID, membershipID,
	).Scan(
		&mr.MembershipID,
		&mr.UserID,
		&mr.OrgID,
		&mr.Email,
		&mr.DisplayName,
		&mr.Role,
		&mr.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	record := mr.toModel()
	return &record, nil
}

func (r *membershipRepository) UpdateRole(ctx context.Context, membershipID, role string) error {
	commandTag, err := r.db.Exec(ctx, `UPDATE memberships SET role = $1 WHERE id = $2`, role, membershipID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *membershipRepository) Create(ctx context.Context, m *model.Membership) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO memberships (id, user_id, org_id, role, created_at) VALUES ($1, $2, $3, $4, $5)",
		m.ID, m.UserID, m.OrgID, m.Role, m.CreatedAt,
	)
	return err
}
