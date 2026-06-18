package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errFolderNameExists is the unexported sentinel exposed via ErrFolderNameExists.
var errFolderNameExists = errors.New("folder name already exists under parent")

// errFolderNotFound is the unexported sentinel exposed via ErrFolderNotFound. It
// lets callers (e.g. EnsureFolderPath) distinguish a genuine "no such folder"
// lookup miss via errors.Is rather than matching error-string substrings.
var errFolderNotFound = errors.New("folder not found")

// errFolderReparentCycle is the unexported sentinel exposed via
// ErrFolderReparentCycle. Reparent returns it when the cycle-checked MoveFolder
// UPDATE affected 0 rows, i.e. the new parent is the folder itself or one of its
// descendants. The usecase maps it to its own ErrFolderCycle (HTTP 400).
var errFolderReparentCycle = errors.New("reparent would create a folder cycle")

// errFolderReparentDepthExceeded is the unexported sentinel exposed via
// ErrFolderReparentDepthExceeded. Reparent returns it when the moved subtree's
// deepest leaf would exceed the supplied depth cap. Because the depth check now
// runs inside the same advisory-locked transaction as the move, the cap is hard
// (no concurrent reparent can change the tree between the check and the write).
var errFolderReparentDepthExceeded = errors.New("reparent would exceed the maximum folder depth")

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type folderRepository struct {
	queries sqldb.Querier
	// pool is retained so Reparent can open a transaction, take a tenant-scoped
	// advisory lock, and run the cycle-checked move plus the depth checks inside
	// one transaction. It is nil when the repository is constructed from a bare
	// DBTX (integration helper), in which case Reparent is unsupported (mirrors
	// documentRepository.ApplySuggestion).
	pool *pgxpool.Pool
}

func NewFolderRepository(db *pgxpool.Pool) FolderRepository {
	return &folderRepository{queries: sqldb.New(db), pool: db}
}

func (r *folderRepository) Create(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	err := r.queries.CreateFolder(ctx, sqldb.CreateFolderParams{
		ID:        folder.ID,
		TenantID:  folder.TenantID,
		OrgID:     folder.OrgID,
		ParentID:  folder.ParentID,
		Name:      folder.Name,
		CreatedBy: folder.CreatedBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return errFolderNameExists
		}
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (r *folderRepository) GetByParentName(ctx context.Context, tenantID, orgID string, parentID *string, name string) (*model.Folder, error) {
	folder, err := r.queries.GetFolderByParentName(ctx, sqldb.GetFolderByParentNameParams{
		TenantID: tenantID,
		OrgID:    orgID,
		ParentID: parentID,
		Name:     name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errFolderNotFound
		}
		return nil, fmt.Errorf("failed to get folder by name: %w", err)
	}
	modelFolder := toModelFolder(folder)
	return &modelFolder, nil
}

// Reparent reparents a folder under a new parent (NULL parentID moves it to
// root) atomically AND race-safely. It is the ONLY sanctioned path for moving a
// folder; the raw MoveFolder SQL must not be called from anywhere else, because
// the cycle/depth safety of a reparent depends entirely on the serialization
// established here.
//
// Cycle-safety mechanism (advisory lock): the cycle check inside the MoveFolder
// UPDATE gives statement atomicity, NOT isolation against a concurrent statement
// on a disjoint row. Under READ COMMITTED two concurrent opposite reparents
// (move A under B in tx1, move B under A in tx2) could each see the other's
// pre-commit tree, each pass its own in-statement descendants check, and both
// commit — creating a cycle. To prevent this we open one transaction and FIRST
// take a transaction-scoped advisory lock keyed on the tenant. All reparents
// within a tenant therefore serialize: the second waiter only proceeds after the
// first commits, so its descendants CTE sees the committed tree and its
// cycle-checked UPDATE correctly returns 0 rows (-> ErrFolderReparentCycle).
//
// The depth check (target-parent ancestors + moved-subtree height) runs inside
// the same locked transaction, making the depth cap HARD: no concurrent reparent
// can change the tree between the check and the move.
//
// rows-affected == 0 from the cycle-checked UPDATE maps to
// ErrFolderReparentCycle. A non-nil parentID that exceeds maxDepth maps to
// ErrFolderReparentDepthExceeded. Move-to-root (NULL parent) skips the
// cycle/depth guards and always succeeds for an existing row.
func (r *folderRepository) Reparent(ctx context.Context, tenantID, orgID, folderID string, parentID *string, maxDepth int) error {
	if r.pool == nil {
		return fmt.Errorf("reparent requires a transactional repository")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin reparent transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Serialize all reparents for this tenant. The transaction-scoped advisory
	// lock is released automatically on commit/rollback. hashtext yields int4;
	// cast to bigint to select the single-key pg_advisory_xact_lock overload.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1)::bigint)", tenantID); err != nil {
		return fmt.Errorf("failed to acquire reparent lock: %w", err)
	}

	q := sqldb.New(tx)

	// Depth is a HARD cap inside the lock: the moved folder's resulting depth is
	// the target parent's depth (its ancestor count, inclusive) plus the height
	// of the moved subtree. Only enforced for a non-NULL target (root is depth 1).
	if parentID != nil && *parentID != "" {
		ancestors, err := q.GetFolderAncestorIDs(ctx, sqldb.GetFolderAncestorIDsParams{
			FolderID: *parentID,
			TenantID: tenantID,
			OrgID:    orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to compute target ancestors: %w", err)
		}
		height, err := q.GetFolderSubtreeHeight(ctx, sqldb.GetFolderSubtreeHeightParams{
			FolderID: folderID,
			TenantID: tenantID,
			OrgID:    orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to compute moved subtree height: %w", err)
		}
		if len(ancestors)+int(height) > maxDepth {
			return errFolderReparentDepthExceeded
		}
	}

	rows, err := q.MoveFolder(ctx, sqldb.MoveFolderParams{
		NewParent: parentID,
		ID:        folderID,
		TenantID:  tenantID,
		OrgID:     orgID,
	})
	if err != nil {
		return fmt.Errorf("failed to move folder: %w", err)
	}
	if rows == 0 {
		// The cycle-checked UPDATE rejected the new parent as the folder itself or
		// a descendant of it. (A NULL parent on an existing row always affects 1.)
		return errFolderReparentCycle
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit reparent transaction: %w", err)
	}
	return nil
}

func (r *folderRepository) GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Folder, error) {
	folder, err := r.queries.GetFolderByID(ctx, sqldb.GetFolderByIDParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	modelFolder := toModelFolder(folder)
	return &modelFolder, nil
}

func (r *folderRepository) ListByParent(ctx context.Context, tenantID, orgID, parentID string) ([]model.Folder, error) {
	folders, err := r.queries.ListFoldersByParent(ctx, sqldb.ListFoldersByParentParams{
		TenantID: tenantID,
		OrgID:    orgID,
		ParentID: &parentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	out := make([]model.Folder, 0, len(folders))
	for _, f := range folders {
		out = append(out, folderListRow(f.ID, f.TenantID, f.OrgID, f.ParentID, f.Name, f.CreatedBy, f.CreatedAt.Time))
	}
	return out, nil
}

func (r *folderRepository) ListRoot(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	folders, err := r.queries.ListRootFolders(ctx, sqldb.ListRootFoldersParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list root folders: %w", err)
	}
	out := make([]model.Folder, 0, len(folders))
	for _, f := range folders {
		out = append(out, folderListRow(f.ID, f.TenantID, f.OrgID, f.ParentID, f.Name, f.CreatedBy, f.CreatedAt.Time))
	}
	return out, nil
}

func (r *folderRepository) ListAll(ctx context.Context, tenantID, orgID string) ([]model.Folder, error) {
	folders, err := r.queries.ListAllFolders(ctx, sqldb.ListAllFoldersParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all folders: %w", err)
	}
	out := make([]model.Folder, 0, len(folders))
	for _, f := range folders {
		out = append(out, folderListRow(f.ID, f.TenantID, f.OrgID, f.ParentID, f.Name, f.CreatedBy, f.CreatedAt.Time))
	}
	return out, nil
}

func (r *folderRepository) Update(ctx context.Context, folder *model.Folder) error {
	if folder == nil {
		return fmt.Errorf("folder is nil")
	}
	err := r.queries.UpdateFolder(ctx, sqldb.UpdateFolderParams{
		Name:     folder.Name,
		ParentID: folder.ParentID,
		ID:       folder.ID,
		TenantID: folder.TenantID,
		OrgID:    folder.OrgID,
	})
	if err != nil {
		return fmt.Errorf("failed to update folder: %w", err)
	}
	return nil
}

func (r *folderRepository) Delete(ctx context.Context, tenantID, orgID, id string) error {
	rowsAffected, err := r.queries.DeleteFolder(ctx, sqldb.DeleteFolderParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}

// GetIndex returns the folder's optional markdown index content. A nil pointer
// means the folder has no index content (or, for a missing folder, ErrNoRows is
// mapped to errFolderNotFound).
func (r *folderRepository) GetIndex(ctx context.Context, tenantID, orgID, id string) (*string, error) {
	content, err := r.queries.GetFolderIndex(ctx, sqldb.GetFolderIndexParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errFolderNotFound
		}
		return nil, fmt.Errorf("failed to get folder index: %w", err)
	}
	return content, nil
}

// SetIndex updates the folder's index content (a nil content clears it). It
// returns errFolderNotFound when no folder matched the tenant/org/id.
func (r *folderRepository) SetIndex(ctx context.Context, tenantID, orgID, id string, content *string) error {
	rows, err := r.queries.SetFolderIndex(ctx, sqldb.SetFolderIndexParams{
		IndexContent: content,
		ID:           id,
		TenantID:     tenantID,
		OrgID:        orgID,
	})
	if err != nil {
		return fmt.Errorf("failed to set folder index: %w", err)
	}
	if rows == 0 {
		return errFolderNotFound
	}
	return nil
}

// folderListRow maps the scalar columns shared by the folder LIST queries into
// a model.Folder. IndexContent is intentionally left nil: list endpoints never
// expose the folder overview (it is served only by the dedicated, visibility-
// checked folder index endpoint).
func folderListRow(id, tenantID, orgID string, parentID *string, name string, createdBy *string, createdAt time.Time) model.Folder {
	return model.Folder{
		ID:        id,
		TenantID:  tenantID,
		OrgID:     orgID,
		ParentID:  parentID,
		Name:      name,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
	}
}

func toModelFolder(folder sqldb.Folder) model.Folder {
	return model.Folder{
		ID:           folder.ID,
		TenantID:     folder.TenantID,
		OrgID:        folder.OrgID,
		ParentID:     folder.ParentID,
		Name:         folder.Name,
		CreatedBy:    folder.CreatedBy,
		CreatedAt:    folder.CreatedAt.Time,
		IndexContent: folder.IndexContent,
	}
}
