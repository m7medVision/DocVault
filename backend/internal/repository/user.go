package repository

import (
	"context"
	"fmt"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	UpdateProfile(ctx context.Context, userID, displayName, locale string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error
	IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error)
}

type userRepository struct {
	queries sqldb.Querier
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{queries: sqldb.New(db)}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (r *userRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
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

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
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

func (r *userRepository) UpdateProfile(ctx context.Context, userID, displayName, locale string) error {
	return r.queries.UpdateUserProfile(ctx, sqldb.UpdateUserProfileParams{
		DisplayName: displayName,
		Locale:      locale,
		ID:          userID,
	})
}

func (r *userRepository) UpdateEmail(ctx context.Context, userID, email string) error {
	return r.queries.UpdateUserEmail(ctx, sqldb.UpdateUserEmailParams{
		Email: email,
		ID:    userID,
	})
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, sqldb.UpdateUserPasswordParams{
		PasswordHash: &passwordHash,
		ID:           userID,
	})
}

func (r *userRepository) UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error {
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

func (r *userRepository) IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error) {
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
