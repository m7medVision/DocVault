package repository

import (
	"context"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, display_name, locale, email_verified,
		 failed_login_attempts, locked_until, last_login_at, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Locale,
		&user.EmailVerified,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.LastLoginAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, email, display_name, locale, email_verified, last_login_at, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.DisplayName,
		&user.Locale,
		&user.EmailVerified,
		&user.LastLoginAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, tenant_id, email, password_hash, display_name, locale, email_verified, failed_login_attempts, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		u.ID, u.TenantID, u.Email, u.PasswordHash, u.DisplayName, u.Locale, u.EmailVerified, u.FailedLoginAttempts, u.CreatedAt,
	)
	return err
}

func (r *userRepository) UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET failed_login_attempts = $1, locked_until = $2 WHERE id = $3",
		attempts, lockedUntil, userID,
	)
	return err
}
