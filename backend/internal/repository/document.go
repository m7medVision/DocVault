// Package repository handles data access to PostgreSQL and MinIO.
package repository

import (
	"context"
	"errors"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/document"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDocumentNotFound is returned when no document matches the tenant/org/id.
// Callers detect it via errors.Is rather than matching error-string substrings.
var ErrDocumentNotFound = errors.New("document not found")

// ErrDocumentTitleExists is returned when a write collides with the per-folder
// unique document-title constraint (idx_documents_unique_title).
var ErrDocumentTitleExists = errors.New("a document with this title already exists in the folder")

// isTitleConflict reports whether err is specifically the per-folder unique
// document-title violation, distinguishing it from any other unique-constraint
// violation (SQLSTATE 23505).
func isTitleConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "idx_documents_unique_title"
}

// documentRepository handles document data access.
type documentRepository struct {
	queries sqldb.Querier
	// pool is retained so operations that must be atomic (e.g. ApplySuggestion)
	// can open a transaction and run several statements through tx-scoped
	// queries. It is nil when the repository is constructed from a bare DBTX
	// (integration tests), in which case transactional methods are unsupported.
	pool *pgxpool.Pool
}

// NewDocumentRepository creates a new DocumentRepository.
func NewDocumentRepository(db *pgxpool.Pool) DocumentRepository {
	return &documentRepository{queries: sqldb.New(db), pool: db}
}

// Create creates a new document.
func (r *documentRepository) Create(ctx context.Context, doc *model.Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	err := r.queries.CreateDocument(ctx, sqldb.CreateDocumentParams{
		ID:              doc.ID,
		TenantID:        doc.TenantID,
		OrgID:           doc.OrgID,
		FolderID:        doc.FolderID,
		OwnerID:         doc.OwnerID,
		Title:           doc.Title,
		DocType:         sqldb.DocumentType(doc.DocType),
		Status:          sqldb.DocumentStatus(doc.Status),
		Language:        doc.Language,
		ProcessingStage: doc.ProcessingStage,
	})
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

// GetByID retrieves a document by ID.
func (r *documentRepository) GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Document, error) {
	doc, err := r.queries.GetDocumentByID(ctx, sqldb.GetDocumentByIDParams{
		ID:       id,
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	modelDoc := model.Document{
		ID:                   doc.ID,
		TenantID:             doc.TenantID,
		OrgID:                doc.OrgID,
		FolderID:             doc.FolderID,
		OwnerID:              doc.OwnerID,
		Title:                doc.Title,
		DocType:              string(doc.DocType),
		Status:               model.DocumentStatus(doc.Status),
		Language:             doc.Language,
		ProcessingStage:      doc.ProcessingStage,
		ProcessingError:      doc.ProcessingError,
		SuggestedFolderName:  doc.SuggestedFolderName,
		SuggestedFilename:    doc.SuggestedFilename,
		SuggestionConfidence: doc.SuggestionConfidence,
		SuggestionCreateNew:  doc.SuggestionCreateNew,
		CreatedAt:            doc.CreatedAt.Time,
	}
	return &modelDoc, nil
}

// ListDocumentsQuery contains filters for listing documents.
type ListDocumentsQuery struct {
	TenantID string
	OrgID    string
	DocType  string
	FolderID string
	Status   model.DocumentStatus
	Language string
	Cursor   string
	Limit    int
}

// List lists documents with filters and cursor pagination.
func (r *documentRepository) List(ctx context.Context, q *ListDocumentsQuery) ([]model.Document, *string, error) {
	if q.Limit <= 0 {
		q.Limit = 20
	}

	var docType *string
	if q.DocType != "" {
		docType = &q.DocType
	}
	var folderID *string
	if q.FolderID != "" {
		folderID = &q.FolderID
	}
	var status *string
	if q.Status != "" {
		value := string(q.Status)
		status = &value
	}
	var language *string
	if q.Language != "" {
		language = &q.Language
	}
	var cursorID *string
	if q.Cursor != "" {
		cursorID = &q.Cursor
	}

	documentRows, err := r.queries.ListDocuments(ctx, sqldb.ListDocumentsParams{
		TenantIDArg: q.TenantID,
		OrgIDArg:    q.OrgID,
		DocType:     docType,
		FolderID:    folderID,
		Status:      status,
		Language:    language,
		CursorID:    cursorID,
		LimitCount:  int32(q.Limit + 1),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list documents: %w", err)
	}

	docs := make([]model.Document, 0, len(documentRows))
	for _, doc := range documentRows {
		docs = append(docs, model.Document{
			ID:        doc.ID,
			TenantID:  doc.TenantID,
			OrgID:     doc.OrgID,
			FolderID:  doc.FolderID,
			OwnerID:   doc.OwnerID,
			Title:     doc.Title,
			DocType:   string(doc.DocType),
			Status:    model.DocumentStatus(doc.Status),
			Language:  doc.Language,
			CreatedAt: doc.CreatedAt.Time,
		})
	}

	var cursor *string
	if len(docs) > q.Limit {
		docs = docs[:q.Limit]
		c := docs[len(docs)-1].ID
		cursor = &c
	}

	return docs, cursor, nil
}

// Update updates a document.
func (r *documentRepository) Update(ctx context.Context, doc *model.Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	err := r.queries.UpdateDocument(ctx, sqldb.UpdateDocumentParams{
		Title:    doc.Title,
		DocType:  sqldb.DocumentType(doc.DocType),
		Status:   sqldb.DocumentStatus(doc.Status),
		Language: doc.Language,
		FolderID: doc.FolderID,
		ID:       doc.ID,
		TenantID: doc.TenantID,
		OrgID:    doc.OrgID,
	})
	if err != nil {
		if isTitleConflict(err) {
			return ErrDocumentTitleExists
		}
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

// Delete removes a document record.
func (r *documentRepository) Delete(ctx context.Context, tenantID, orgID, id, actorID string) error {
	rowsAffected, err := r.queries.DeleteDocument(ctx, sqldb.DeleteDocumentParams{ID: id, TenantID: tenantID, OrgID: orgID})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

// CreateVersion creates a new document version.
func (r *documentRepository) CreateVersion(ctx context.Context, version *model.DocumentVersion) error {
	if version == nil {
		return fmt.Errorf("version is nil")
	}
	err := r.queries.CreateDocumentVersion(ctx, sqldb.CreateDocumentVersionParams{
		ID:            version.ID,
		DocumentID:    version.DocumentID,
		VersionNumber: int32(version.VersionNumber),
		StorageKey:    version.StorageKey,
		MimeType:      version.MimeType,
		FileSize:      version.FileSize,
		UploadedBy:    version.UploadedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create document version: %w", err)
	}
	return nil
}

// GetVersions retrieves all versions of a document.
func (r *documentRepository) GetVersions(ctx context.Context, tenantID, documentID string) ([]model.DocumentVersion, error) {
	versions, err := r.queries.GetDocumentVersions(ctx, sqldb.GetDocumentVersionsParams{
		DocumentID: documentID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	modelVersions := make([]model.DocumentVersion, 0, len(versions))
	for _, version := range versions {
		modelVersions = append(modelVersions, model.DocumentVersion{
			ID:            version.ID,
			DocumentID:    version.DocumentID,
			VersionNumber: int(version.VersionNumber),
			StorageKey:    version.StorageKey,
			MimeType:      version.MimeType,
			FileSize:      version.FileSize,
			UploadedBy:    version.UploadedBy,
			CreatedAt:     version.CreatedAt.Time,
		})
	}
	return modelVersions, nil
}

// CreatePage creates a new document page.
func (r *documentRepository) CreatePage(ctx context.Context, page *model.DocumentPage) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	err := r.queries.CreateDocumentPage(ctx, sqldb.CreateDocumentPageParams{
		ID:             page.ID,
		DocumentID:     page.DocumentID,
		VersionID:      page.VersionID,
		PageNumber:     int32(page.PageNumber),
		OcrText:        page.OCRText,
		TranslatedText: page.TranslatedText,
		Confidence:     page.Confidence,
		OcrModel:       page.OCRModel,
	})
	if err != nil {
		return fmt.Errorf("failed to create document page: %w", err)
	}
	return nil
}

// GetPages retrieves all pages of a document.
func (r *documentRepository) GetPages(ctx context.Context, tenantID, documentID string) ([]model.DocumentPage, error) {
	pages, err := r.queries.GetDocumentPages(ctx, sqldb.GetDocumentPagesParams{
		DocumentID: documentID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}
	modelPages := make([]model.DocumentPage, 0, len(pages))
	for _, page := range pages {
		modelPages = append(modelPages, model.DocumentPage{
			ID:             page.ID,
			DocumentID:     page.DocumentID,
			VersionID:      page.VersionID,
			PageNumber:     int(page.PageNumber),
			OCRText:        page.OcrText,
			TranslatedText: page.TranslatedText,
			Confidence:     page.Confidence,
			OCRModel:       page.OcrModel,
			CreatedAt:      page.CreatedAt.Time,
		})
	}
	return modelPages, nil
}

// SetMetadata sets document metadata.
func (r *documentRepository) SetMetadata(ctx context.Context, tenantID string, metadata *model.DocumentMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}
	err := r.queries.SetDocumentMetadata(ctx, sqldb.SetDocumentMetadataParams{
		ID:             metadata.ID,
		DocumentID:     metadata.DocumentID,
		Key:            metadata.Key,
		ExtractedValue: metadata.ExtractedValue,
		CorrectedValue: metadata.CorrectedValue,
		CorrectedBy:    metadata.CorrectedBy,
		CorrectedAt:    timestamptzFromTimePtr(metadata.CorrectedAt),
	})
	if err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}
	return nil
}

// GetMetadata retrieves document metadata.
func (r *documentRepository) GetMetadata(ctx context.Context, tenantID, documentID string) ([]model.DocumentMetadata, error) {
	metadata, err := r.queries.GetDocumentMetadata(ctx, sqldb.GetDocumentMetadataParams{
		DocumentID: documentID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	modelMetadata := make([]model.DocumentMetadata, 0, len(metadata))
	for _, item := range metadata {
		modelMetadata = append(modelMetadata, model.DocumentMetadata{
			ID:             item.ID,
			DocumentID:     item.DocumentID,
			Key:            item.Key,
			ExtractedValue: item.ExtractedValue,
			CorrectedValue: item.CorrectedValue,
			CorrectedBy:    item.CorrectedBy,
			CorrectedAt:    timePtrFromTimestamptz(item.CorrectedAt),
			CreatedAt:      item.CreatedAt.Time,
		})
	}
	return modelMetadata, nil
}

// UpdateMetadataField updates a single metadata field with corrected value.
func (r *documentRepository) UpdateMetadataField(ctx context.Context, tenantID, documentID, key, correctedValue, correctedBy string) error {
	rowsAffected, err := r.queries.UpdateDocumentMetadataField(ctx, sqldb.UpdateDocumentMetadataFieldParams{
		CorrectedValue: &correctedValue,
		CorrectedBy:    &correctedBy,
		DocumentID:     documentID,
		Key:            key,
		TenantID:       tenantID,
	})
	if err != nil {
		return fmt.Errorf("failed to update metadata field: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("metadata key not found: %s", key)
	}
	return nil
}

// GetFullDocument retrieves a document with all its versions, pages, and metadata.
func (r *documentRepository) GetFullDocument(ctx context.Context, tenantID, orgID, documentID string) (*model.Document, []model.DocumentVersion, []model.DocumentMetadata, error) {
	doc, err := r.GetByID(ctx, tenantID, orgID, documentID)
	if err != nil {
		return nil, nil, nil, err
	}

	versions, err := r.GetVersions(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, nil, err
	}

	metadata, err := r.GetMetadata(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, nil, err
	}

	return doc, versions, metadata, nil
}

func (r *documentRepository) UpdateProcessingFields(ctx context.Context, tenantID, documentID string, stage *string, errMsg *string) error {
	execErr := r.queries.UpdateDocumentProcessingFields(ctx, sqldb.UpdateDocumentProcessingFieldsParams{
		ProcessingStage: stage,
		ProcessingError: errMsg,
		ID:              documentID,
		TenantID:        tenantID,
	})
	if execErr != nil {
		return fmt.Errorf("failed to update processing fields: %w", execErr)
	}
	return nil
}

// ClearSuggestion clears the four folder-suggestion columns and optionally sets
// the processing_stage (e.g. "completed" on accept; nil leaves it unchanged).
func (r *documentRepository) ClearSuggestion(ctx context.Context, tenantID, orgID, documentID string, stage *string) error {
	err := r.queries.ClearDocumentSuggestion(ctx, sqldb.ClearDocumentSuggestionParams{
		ProcessingStage: stage,
		ID:              documentID,
		TenantID:        tenantID,
		OrgID:           orgID,
	})
	if err != nil {
		return fmt.Errorf("failed to clear document suggestion: %w", err)
	}
	return nil
}

// ApplySuggestion files a document and clears its suggestion atomically: the
// folder/title update and the suggestion-column clear (with the given stage)
// run inside a single transaction so the document can never be left filed but
// with a stale, uncleared suggestion (or vice versa) if one statement fails.
// The caller passes the already-resolved document (with FolderID/Title set to
// their accepted values); only the columns written by UpdateDocument are
// modified. On any error the transaction is rolled back.
func (r *documentRepository) ApplySuggestion(ctx context.Context, doc *model.Document, stage *string) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	if r.pool == nil {
		return fmt.Errorf("apply suggestion requires a transactional repository")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := sqldb.New(tx)

	if err := q.UpdateDocument(ctx, sqldb.UpdateDocumentParams{
		Title:    doc.Title,
		DocType:  sqldb.DocumentType(doc.DocType),
		Status:   sqldb.DocumentStatus(doc.Status),
		Language: doc.Language,
		FolderID: doc.FolderID,
		ID:       doc.ID,
		TenantID: doc.TenantID,
		OrgID:    doc.OrgID,
	}); err != nil {
		if isTitleConflict(err) {
			return ErrDocumentTitleExists
		}
		return fmt.Errorf("failed to apply suggestion update: %w", err)
	}

	if err := q.ClearDocumentSuggestion(ctx, sqldb.ClearDocumentSuggestionParams{
		ProcessingStage: stage,
		ID:              doc.ID,
		TenantID:        doc.TenantID,
		OrgID:           doc.OrgID,
	}); err != nil {
		return fmt.Errorf("failed to clear suggestion in transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit suggestion transaction: %w", err)
	}
	return nil
}

func (r *documentRepository) GetStats(ctx context.Context, tenantID, orgID string) (*model.DocumentStats, error) {
	stats, err := r.queries.GetDocumentStats(ctx, sqldb.GetDocumentStatsParams{TenantIDArg: tenantID, OrgIDArg: orgID})
	if err != nil {
		return nil, fmt.Errorf("getting document stats: %w", err)
	}

	return &model.DocumentStats{
		TotalDocuments:    stats.TotalDocuments,
		PendingDocuments:  stats.PendingDocuments,
		CompletedThisWeek: stats.CompletedThisWeek,
		StorageUsedBytes:  stats.StorageUsedBytes,
	}, nil
}
