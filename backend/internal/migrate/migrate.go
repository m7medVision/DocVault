// Package migrate runs embedded SQL migrations via goose.
package migrate

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver
	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

// Run applies all pending migrations against the database at dbURL.
func Run(ctx context.Context, dbURL string) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	if err := goose.UpContext(ctx, db, "sql"); err != nil {
		return wrapUpError(ctx, db, err)
	}

	return nil
}

// Status prints the current migration status to stdout.
func Status(ctx context.Context, dbURL string) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	return goose.StatusContext(ctx, db, "sql")
}

func wrapUpError(ctx context.Context, db *sql.DB, err error) error {
	// 42P07 = duplicate_table, 42710 = duplicate_object (e.g. a pre-existing ENUM
	// type). Both mean a migration is creating something that already exists, i.e.
	// the database was migrated before and goose's version table is out of sync —
	// the classic "reused or partially-reset dev database" symptom.
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || (pgErr.Code != "42P07" && pgErr.Code != "42710") {
		return fmt.Errorf("goose up: %w", err)
	}

	tables, total, tableErr := existingPublicTables(ctx, db, 8)
	if tableErr != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return fmt.Errorf(
		"goose up: database already contains conflicting objects (tables/types; existing tables: %s), so DocVault migrations are running against a reused or non-empty database; original error: %w. Point DATABASE_URL at a fresh database, or reset the local Postgres volume with `just dev-clean && just dev-up` if you do not need the current data",
		formatTableSummary(tables, total),
		err,
	)
}

func existingPublicTables(ctx context.Context, db *sql.DB, limit int) ([]string, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'goose_db_version'
	`

	var total int
	if err := db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQuery = `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'goose_db_version'
		ORDER BY tablename
		LIMIT $1
	`

	rows, err := db.QueryContext(ctx, listQuery, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tables := make([]string, 0, limit)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, 0, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tables, total, nil
}

func formatTableSummary(tables []string, total int) string {
	if total == 0 {
		return "none"
	}

	summary := strings.Join(tables, ", ")
	if summary == "" {
		return fmt.Sprintf("%d total", total)
	}

	if total > len(tables) {
		return fmt.Sprintf("%s, ... (%d total)", summary, total)
	}

	return fmt.Sprintf("%s (%d total)", summary, total)
}
