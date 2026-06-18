//go:build integration

package repository

import sqldb "github.com/docvault/backend/internal/db"

// NewFolderRepositoryFromDBTX builds a FolderRepository bound to any sqldb.DBTX
// (a *pgxpool.Pool or a pgx.Tx). It exists only under the `integration` build
// tag so DB-backed tests in the external repository_test package can wire the
// real repository to a rolled-back transaction without depending on a live
// pool. It is never compiled into production binaries.
func NewFolderRepositoryFromDBTX(db sqldb.DBTX) FolderRepository {
	return &folderRepository{queries: sqldb.New(db)}
}
