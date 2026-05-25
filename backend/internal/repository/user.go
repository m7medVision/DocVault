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
	UpdateProfile(ctx context.Context, userID, displayName, locale string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *string) error
	IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error)
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
		`SELECT id, tenant_id, email, COALESCE(password_hash, ''), display_name, locale, email_verified,
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
		`SELECT id, tenant_id, email, COALESCE(password_hash, ''), display_name, locale, email_verified, last_login_at, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
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

func (r *userRepository) UpdateProfile(ctx context.Context, userID, displayName, locale string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET display_name = $1, locale = $2 WHERE id = $3",
		displayName, locale, userID,
	)
	return err
}

func (r *userRepository) UpdateEmail(ctx context.Context, userID, email string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET email = $1, email_verified = FALSE WHERE id = $2",
		email, userID,
	)
	return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET password_hash = $1, failed_login_attempts = 0, locked_until = NULL WHERE id = $2",
		passwordHash, userID,
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

func (r *userRepository) IsEmailTakenByOther(ctx context.Context, email, excludeUserID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id <> $2)",
		email, excludeUserID,
	).Scan(&exists)
	return exists, err
}
