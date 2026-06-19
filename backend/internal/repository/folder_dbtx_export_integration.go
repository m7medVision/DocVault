//go:build integration

package repository

import (
	sqldb "github.com/docvault/backend/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewFolderRepositoryFromDBTX builds a FolderRepository bound to any sqldb.DBTX
// (a *pgxpool.Pool or a pgx.Tx). It exists only under the `integration` build
// tag so DB-backed tests in the external repository_test package can wire the
// real repository to a rolled-back transaction without depending on a live
// pool. It is never compiled into production binaries.
//
// The resulting repository has a nil pool, so Reparent (which needs to open its
// own advisory-locked transaction) is unsupported on it. Tests that exercise
// Reparent must use NewFolderRepositoryFromPool with real committed rows.
func NewFolderRepositoryFromDBTX(db sqldb.DBTX) FolderRepository {
	return &folderRepository{queries: sqldb.New(db)}
}

// NewFolderRepositoryFromPool builds a pool-backed FolderRepository identical to
// the production one, exposed under the `integration` build tag so DB-backed
// tests can exercise Reparent (which opens its own advisory-locked transaction
// against the pool and commits). Never compiled into production binaries.
func NewFolderRepositoryFromPool(pool *pgxpool.Pool) FolderRepository {
	return &folderRepository{queries: sqldb.New(pool), pool: pool}
}
