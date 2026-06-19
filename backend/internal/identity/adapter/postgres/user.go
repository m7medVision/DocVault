// Package postgres is the identity bounded context's data-access adapter. It
// wraps the shared sqlc Queries and maps rows to the identity domain model for
// users, memberships, and transactional workspace registration.
package postgres

import (
	"context"
	"fmt"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles user data access. It satisfies the
// repository.UserRepository contract; the composition root binds this concrete
// type to that interface.
type UserRepository struct {
	queries sqldb.Querier
}

// NewUserRepository creates a postgres-backed user repository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: sqldb.New(db)}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, sqldb.FindUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:                  row.ID,
		TenantID:            row.TenantID,
		Email:               row.Email,
		PasswordHash:        row.PasswordHash,
		DisplayName:         row.DisplayName,
		Locale:              row.Locale,
		EmailVerified:       row.EmailVerified,
		FailedLoginAttempts: int(row.FailedLoginAttempts),
		LockedUntil:         row.LockedUntil,
		LastLoginAt:         row.LastLoginAt,
		CreatedAt:           row.CreatedAt.Time,
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	row, err := r.queries.FindUserByID(ctx, sqldb.FindUserByIDParams{ID: userID})
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:            row.ID,
		TenantID:      row.TenantID,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		DisplayName:   row.DisplayName,
		Locale:        row.Locale,
		EmailVerified: row.EmailVerified,
		LastLoginAt:   row.LastLoginAt,
		CreatedAt:     row.CreatedAt.Time,
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	return r.queries.CreateUser(ctx, sqldb.CreateUserParams{
		ID:                  u.ID,
		TenantID:            u.TenantID,
		Email:               u.Email,
		PasswordHash:        &u.PasswordHash,
		DisplayName:         u.DisplayName,
		Locale:              u.Locale,
		EmailVerified:       u.EmailVerified,
		FailedLoginAttempts: int32(u.FailedLoginAttempts),
		CreatedAt:           pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
	})
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID, displayName, locale string) error {
	return r.queries.UpdateUserProfile(ctx, sqldb.UpdateUserProfileParams{
		DisplayName: displayName,
		Locale:      locale,
		ID:          userID,
	})
}

func (r *UserRepository) UpdateEmail(ctx context.Context, userID, email string) error {
	return r.queries.UpdateUserEmail(ctx, sqldb.UpdateUserEmailParams{
		Email: email,
		ID:    userID,
	})
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, sqldb.UpdateUserPasswordParams{
		PasswordHash: &passwordHash,
		ID:           userID,
	})
}

func (r *UserRepository) UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error {
	parsedLockedUntil, err := parseOptionalTime(lockedUntil)
	if err != nil {
		return err
	}
	return r.queries.UpdateUserFailedLogin(ctx, sqldb.UpdateUserFailedLoginParams{
		FailedLoginAttempts: int32(attempts),
		LockedUntil:         parsedLockedUntil,
		ID:                  userID,
	})
}

func (r *UserRepository) IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error) {
	return r.queries.IsEmailTakenByOther(ctx, sqldb.IsEmailTakenByOtherParams{
		Email: email,
		ID:    excludeUserID,
	})
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, fmt.Errorf("parse locked_until: %w", err)
	}
	return &parsed, nil
}
