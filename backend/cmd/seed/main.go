// Command seed populates the dev database with a representative set of tenants,
// organizations, and users across every role.
//
// It mirrors the registration flow in internal/transport/http/auth_register.go so
// seeded users are indistinguishable from real ones: tenant -> org -> user (bcrypt
// hash) -> membership, then Casbin policy + role binding via authz.EnsureTenantRoleAccess.
//
// IDs are derived deterministically from a stable key (uuid v5), so the seeder is
// idempotent: re-running maps to the same rows and already-seeded users are skipped.
//
// Run from the backend/ directory (via `just db-seed`) so the relative
// internal/authz/model.conf path resolves, exactly like the API does.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/config"
	"github.com/docvault/backend/internal/database"
	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/migrate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seedPassword is shared by every seeded user. It satisfies
// auth.ValidatePasswordStrength (upper, lower, digit, special, >= 8 chars).
const seedPassword = "Passw0rd!"

// seedNamespace anchors the deterministic (uuid v5) IDs the seeder generates, so
// the same fixture key always maps to the same database row across runs.
var seedNamespace = uuid.MustParse("d0c5eed0-0000-4000-8000-5eed5eed5eed")

// user is a single seeded identity within a tenant's organization.
type user struct {
	role        string // authz.Role* / membership role
	displayName string
	locale      string // "en" or "ar"
}

// tenant is a workspace seeded with one organization and a spread of users.
type tenant struct {
	slug  string // also the email domain: <local>@<slug>.test
	name  string
	org   string
	plan  string
	users []user
}

// fixture is the data the seeder creates. Emails are derived as
// <role><n>@<slug>.test (the index suffix is only added when a role repeats).
var fixture = []tenant{
	{
		slug: "acme", name: "Acme Legal", org: "Acme Legal LLP", plan: "business",
		users: []user{
			{authz.RoleOwner, "Aisha Owner", "en"},
			{authz.RoleAdmin, "Adam Admin", "en"},
			{authz.RoleMember, "Maya Member", "en"},
			{authz.RoleMember, "Marco Member", "ar"},
			{authz.RoleViewer, "Vera Viewer", "en"},
		},
	},
	{
		slug: "gulf", name: "Gulf Finance", org: "Gulf Finance Holding", plan: "business",
		users: []user{
			{authz.RoleOwner, "Omar Owner", "en"},
			{authz.RoleAdmin, "Amira Admin", "ar"},
			{authz.RoleMember, "Mona Member", "en"},
			{authz.RoleViewer, "Yusuf Viewer", "en"},
		},
	},
	{
		slug: "doha", name: "Doha Admin Office", org: "Doha Admin Office", plan: "personal",
		users: []user{
			{authz.RoleOwner, "Khalid Owner", "ar"},
			{authz.RoleMember, "Sara Member", "ar"},
		},
	},
}

// credential records a successfully ensured login for the final summary table.
type credential struct {
	email, tenant, role, locale string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Apply migrations first so a fresh database can be seeded in one step,
	// mirroring how the API boots (internal/app/api.go).
	if err := migrate.Run(ctx, cfg.DB.URL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	pool, err := database.NewConnection(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	enforcer, err := authz.NewEnforcer(filepath.Join("internal", "authz", "model.conf"), cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("build authz enforcer: %w", err)
	}

	queries := sqldb.New(pool)

	var creds []credential
	for _, t := range fixture {
		tenantCreds, err := seedTenant(ctx, logger, pool, queries, enforcer, t)
		if err != nil {
			return fmt.Errorf("seed tenant %q: %w", t.slug, err)
		}
		creds = append(creds, tenantCreds...)
	}

	printSummary(creds)
	return nil
}

// seedTenant creates one tenant, its organization, and its users + memberships in a
// single transaction, then seeds Casbin policies/role bindings per user after commit
// (the enforcer writes to its own casbin_rule store, outside the tx — same ordering
// as the register handler). Existing rows are detected and skipped for idempotency.
func seedTenant(
	ctx context.Context,
	logger *slog.Logger,
	pool *pgxpool.Pool,
	queries *sqldb.Queries,
	enforcer *casbin.Enforcer,
	t tenant,
) ([]credential, error) {
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	tenantID := deterministicID("tenant:" + t.slug)
	orgID := deterministicID("org:" + t.slug)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := queries.WithTx(tx)

	if err := ensureTenant(ctx, tx, q, tenantID, t, now); err != nil {
		return nil, err
	}
	if err := ensureOrganization(ctx, tx, q, orgID, tenantID, t, now); err != nil {
		return nil, err
	}

	passwordHash, err := auth.HashPassword(seedPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// roleCounts disambiguates repeated roles in the email local-part (member, member2…).
	roleCounts := map[string]int{}
	type ensured struct {
		userID, email, role string
		locale              string
	}
	var ensuredUsers []ensured

	for _, u := range t.users {
		roleCounts[u.role]++
		email := localPart(u.role, roleCounts[u.role]) + "@" + t.slug + ".test"
		userID := deterministicID("user:" + email)

		existing, err := q.FindUserByEmail(ctx, sqldb.FindUserByEmailParams{Email: email})
		switch {
		case err == nil:
			// Already seeded — keep its real ID for the role binding and skip inserts.
			userID = existing.ID
			logger.Info("user exists, skipping insert", "email", email)
		case errors.Is(err, pgx.ErrNoRows):
			if err := createUser(ctx, q, userID, tenantID, email, &passwordHash, u, now); err != nil {
				return nil, fmt.Errorf("create user %q: %w", email, err)
			}
			if err := createMembership(ctx, q, userID, orgID, u.role, now); err != nil {
				return nil, fmt.Errorf("create membership %q: %w", email, err)
			}
			logger.Info("seeded user", "email", email, "role", u.role)
		default:
			return nil, fmt.Errorf("lookup user %q: %w", email, err)
		}

		ensuredUsers = append(ensuredUsers, ensured{userID, email, u.role, u.locale})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Casbin seeding happens after commit, like auth_register.go.
	var creds []credential
	for _, u := range ensuredUsers {
		if err := authz.EnsureTenantRoleAccess(enforcer, u.userID, u.role, tenantID); err != nil {
			return nil, fmt.Errorf("seed authz for %q: %w", u.email, err)
		}
		creds = append(creds, credential{email: u.email, tenant: t.name, role: u.role, locale: u.locale})
	}
	return creds, nil
}

func ensureTenant(ctx context.Context, tx pgx.Tx, q *sqldb.Queries, id string, t tenant, now pgtype.Timestamptz) error {
	exists, err := rowExists(ctx, tx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", id)
	if err != nil {
		return fmt.Errorf("lookup tenant: %w", err)
	}
	if exists {
		return nil
	}
	return q.CreateTenant(ctx, sqldb.CreateTenantParams{ID: id, Name: t.name, Plan: t.plan, CreatedAt: now})
}

func ensureOrganization(ctx context.Context, tx pgx.Tx, q *sqldb.Queries, id, tenantID string, t tenant, now pgtype.Timestamptz) error {
	exists, err := rowExists(ctx, tx, "SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)", id)
	if err != nil {
		return fmt.Errorf("lookup organization: %w", err)
	}
	if exists {
		return nil
	}
	return q.CreateOrganization(ctx, sqldb.CreateOrganizationParams{ID: id, TenantID: tenantID, Name: t.org, CreatedAt: now})
}

// rowExists runs an EXISTS query (no sqlc query exists for tenant/org lookup by ID).
func rowExists(ctx context.Context, tx pgx.Tx, query, arg string) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, query, arg).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func createUser(ctx context.Context, q *sqldb.Queries, id, tenantID, email string, hash *string, u user, now pgtype.Timestamptz) error {
	return q.CreateUser(ctx, sqldb.CreateUserParams{
		ID:                  id,
		TenantID:            tenantID,
		Email:               email,
		PasswordHash:        hash,
		DisplayName:         u.displayName,
		Locale:              u.locale,
		EmailVerified:       true,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
	})
}

func createMembership(ctx context.Context, q *sqldb.Queries, userID, orgID, role string, now pgtype.Timestamptz) error {
	return q.CreateMembership(ctx, sqldb.CreateMembershipParams{
		ID:        deterministicID("membership:" + userID + ":" + orgID),
		UserID:    userID,
		OrgID:     orgID,
		Role:      sqldb.MembershipRole(role),
		CreatedAt: now,
	})
}

// localPart builds the email local-part, suffixing the index only when a role repeats.
func localPart(role string, n int) string {
	if n <= 1 {
		return role
	}
	return fmt.Sprintf("%s%d", role, n)
}

// deterministicID returns a stable uuid v5 derived from key, so re-seeding is idempotent.
func deterministicID(key string) string {
	return uuid.NewSHA1(seedNamespace, []byte(key)).String()
}

func printSummary(creds []credential) {
	fmt.Printf("\nSeeded %d users. Shared password: %s\n\n", len(creds), seedPassword)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tROLE\tLOCALE\tTENANT")
	fmt.Fprintln(w, "-----\t----\t------\t------")
	for _, c := range creds {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.email, c.role, c.locale, c.tenant)
	}
	w.Flush()
	fmt.Println()
}
