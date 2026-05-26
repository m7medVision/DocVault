package identity

import "time"

// Tenant represents a root isolation unit.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

// Organization represents a sub-unit within a tenant.
type Organization struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a user account with internal authentication.
type User struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	DisplayName         string     `json:"display_name"`
	Locale              string     `json:"locale"`
	EmailVerified       bool       `json:"email_verified"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
}

// Membership links a user to an organization with a role.
type Membership struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
