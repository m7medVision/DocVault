-- name: FindUserByEmail :one
SELECT id, tenant_id, email, COALESCE(password_hash, '') AS password_hash, display_name, locale, email_verified,
       failed_login_attempts, locked_until, last_login_at, created_at
FROM users
WHERE email = $1;

-- name: FindUserByID :one
SELECT id, tenant_id, email, COALESCE(password_hash, '') AS password_hash, display_name, locale, email_verified,
       last_login_at, created_at
FROM users
WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (id, tenant_id, email, password_hash, display_name, locale, email_verified, failed_login_attempts, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateUserProfile :exec
UPDATE users
SET display_name = $1, locale = $2
WHERE id = $3;

-- name: UpdateUserEmail :exec
UPDATE users
SET email = $1, email_verified = FALSE
WHERE id = $2;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $1, failed_login_attempts = 0, locked_until = NULL
WHERE id = $2;

-- name: UpdateUserFailedLogin :exec
UPDATE users
SET failed_login_attempts = $1, locked_until = $2
WHERE id = $3;

-- name: UpdateSuccessfulLoginMetadata :exec
UPDATE users
SET failed_login_attempts = 0, locked_until = NULL, last_login_at = $1
WHERE id = $2;

-- name: IsEmailTakenByOther :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id <> $2);
