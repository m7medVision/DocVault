// Package repository handles data access to PostgreSQL and MinIO.
package repository

import (
	"context"
	"fmt"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// documentRepository handles document data access.
type documentRepository struct {
	db *pgxpool.Pool
}

// NewDocumentRepository creates a new DocumentRepository.
func NewDocumentRepository(db *pgxpool.Pool) DocumentRepository {
	return &documentRepository{db: db}
}

// Create creates a new document.
func (r *documentRepository) Create(ctx context.Context, doc *model.Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	query := `
		INSERT INTO documents (id, tenant_id, org_id, folder_id, owner_id, title, doc_type, status, language, processing_stage, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, doc.OrgID, doc.FolderID, doc.OwnerID,
		doc.Title, doc.DocType, doc.Status, doc.Language, doc.ProcessingStage,
	)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

// GetByID retrieves a document by ID.
func (r *documentRepository) GetByID(ctx context.Context, tenantID, orgID, id string) (*model.Document, error) {
	query := `
		SELECT id, tenant_id, org_id, folder_id, owner_id, title, doc_type, status, language,
		       processing_stage, processing_error,
		       suggested_folder_name, suggested_filename, suggestion_confidence, suggestion_create_new,
		       created_at
		FROM documents
		WHERE id = $1 AND tenant_id = $2 AND org_id = $3
	`
	var doc model.Document
	err := r.db.QueryRow(ctx, query, id, tenantID, orgID).Scan(
		&doc.ID, &doc.TenantID, &doc.OrgID, &doc.FolderID, &doc.OwnerID,
		&doc.Title, &doc.DocType, &doc.Status, &doc.Language,
		&doc.ProcessingStage, &doc.ProcessingError,
		&doc.SuggestedFolderName, &doc.SuggestedFilename, &doc.SuggestionConfidence, &doc.SuggestionCreateNew,
		&doc.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return &doc, nil
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

	query := `
		SELECT id, tenant_id, org_id, folder_id, owner_id, title, doc_type, status, language, created_at
		FROM documents
		WHERE tenant_id = $1 AND org_id = $2
	`
	args := []interface{}{q.TenantID, q.OrgID}
	argCount := 2

	if q.DocType != "" {
		argCount++
		query += fmt.Sprintf(" AND doc_type = $%d", argCount)
		args = append(args, q.DocType)
	}
	if q.FolderID != "" {
		argCount++
		query += fmt.Sprintf(" AND folder_id = $%d", argCount)
		args = append(args, q.FolderID)
	}
	if q.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(q.Status))
	}
	if q.Language != "" {
		argCount++
		query += fmt.Sprintf(" AND language = $%d", argCount)
		args = append(args, q.Language)
	}
	if q.Cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM documents WHERE id = $%d)", argCount)
		args = append(args, q.Cursor)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", q.Limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	docs := make([]model.Document, 0)
	for rows.Next() {
		var doc model.Document
		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.OrgID, &doc.FolderID, &doc.OwnerID,
			&doc.Title, &doc.DocType, &doc.Status, &doc.Language, &doc.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, doc)
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
	query := `
		UPDATE documents
		SET title = $1, doc_type = $2, status = $3, language = $4, folder_id = $5
		WHERE id = $6 AND tenant_id = $7 AND org_id = $8
	`
	_, err := r.db.Exec(ctx, query,
		doc.Title, doc.DocType, doc.Status, doc.Language, doc.FolderID,
		doc.ID, doc.TenantID, doc.OrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

// Delete removes a document record.
func (r *documentRepository) Delete(ctx context.Context, tenantID, orgID, id, actorID string) error {
	query := `DELETE FROM documents WHERE id = $1 AND tenant_id = $2 AND org_id = $3`
	result, err := r.db.Exec(ctx, query, id, tenantID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}
	return nil
}

// CreateVersion creates a new document version.
func (r *documentRepository) CreateVersion(ctx context.Context, version *model.DocumentVersion) error {
	if version == nil {
		return fmt.Errorf("version is nil")
	}
	query := `
		INSERT INTO document_versions (id, document_id, version_number, storage_key, mime_type, file_size, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		version.ID, version.DocumentID, version.VersionNumber,
		version.StorageKey, version.MimeType, version.FileSize, version.UploadedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create document version: %w", err)
	}
	return nil
}

// GetVersions retrieves all versions of a document.
func (r *documentRepository) GetVersions(ctx context.Context, tenantID, documentID string) ([]model.DocumentVersion, error) {
	query := `
		SELECT dv.id, dv.document_id, dv.version_number, dv.storage_key, dv.mime_type, dv.file_size, dv.uploaded_by, dv.created_at
		FROM document_versions dv
		JOIN documents d ON dv.document_id = d.id
		WHERE dv.document_id = $1 AND d.tenant_id = $2
		ORDER BY dv.version_number DESC
	`
	rows, err := r.db.Query(ctx, query, documentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var versions []model.DocumentVersion
	for rows.Next() {
		var v model.DocumentVersion
		if err := rows.Scan(
			&v.ID, &v.DocumentID, &v.VersionNumber, &v.StorageKey,
			&v.MimeType, &v.FileSize, &v.UploadedBy, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// CreatePage creates a new document page.
func (r *documentRepository) CreatePage(ctx context.Context, page *model.DocumentPage) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	query := `
		INSERT INTO document_pages (id, document_id, version_id, page_number, ocr_text, translated_text, confidence, ocr_model, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		page.ID, page.DocumentID, page.VersionID, page.PageNumber,
		page.OCRText, page.TranslatedText, page.Confidence, page.OCRModel,
	)
	if err != nil {
		return fmt.Errorf("failed to create document page: %w", err)
	}
	return nil
}

// GetPages retrieves all pages of a document.
func (r *documentRepository) GetPages(ctx context.Context, tenantID, documentID string) ([]model.DocumentPage, error) {
	query := `
		SELECT p.id, p.document_id, p.version_id, p.page_number, p.ocr_text, p.translated_text, p.confidence, p.ocr_model, p.created_at
		FROM document_pages p
		JOIN documents d ON d.id = p.document_id
		WHERE p.document_id = $1 AND d.tenant_id = $2
		ORDER BY page_number ASC
	`
	rows, err := r.db.Query(ctx, query, documentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}
	defer rows.Close()

	var pages []model.DocumentPage
	for rows.Next() {
		var p model.DocumentPage
		if err := rows.Scan(
			&p.ID, &p.DocumentID, &p.VersionID, &p.PageNumber,
			&p.OCRText, &p.TranslatedText, &p.Confidence, &p.OCRModel, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan page: %w", err)
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// SetMetadata sets document metadata.
func (r *documentRepository) SetMetadata(ctx context.Context, tenantID string, metadata *model.DocumentMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}
	query := `
		INSERT INTO document_metadata (id, document_id, key, extracted_value, corrected_value, corrected_by, corrected_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (document_id, key) DO UPDATE SET
			extracted_value = EXCLUDED.extracted_value,
			corrected_value = EXCLUDED.corrected_value,
			corrected_by = EXCLUDED.corrected_by,
			corrected_at = EXCLUDED.corrected_at
	`
	_, err := r.db.Exec(ctx, query,
		metadata.ID, metadata.DocumentID, metadata.Key,
		metadata.ExtractedValue, metadata.CorrectedValue,
		metadata.CorrectedBy, metadata.CorrectedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}
	return nil
}

// GetMetadata retrieves document metadata.
func (r *documentRepository) GetMetadata(ctx context.Context, tenantID, documentID string) ([]model.DocumentMetadata, error) {
	query := `
		SELECT m.id, m.document_id, m.key, m.extracted_value, m.corrected_value, m.corrected_by, m.corrected_at, m.created_at
		FROM document_metadata m
		JOIN documents d ON d.id = m.document_id
		WHERE m.document_id = $1 AND d.tenant_id = $2
		ORDER BY key ASC
	`
	rows, err := r.db.Query(ctx, query, documentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer rows.Close()

	var metadata []model.DocumentMetadata
	for rows.Next() {
		var m model.DocumentMetadata
		if err := rows.Scan(
			&m.ID, &m.DocumentID, &m.Key,
			&m.ExtractedValue, &m.CorrectedValue,
			&m.CorrectedBy, &m.CorrectedAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan metadata: %w", err)
		}
		metadata = append(metadata, m)
	}
	return metadata, nil
}

// UpdateMetadataField updates a single metadata field with corrected value.
func (r *documentRepository) UpdateMetadataField(ctx context.Context, tenantID, documentID, key, correctedValue, correctedBy string) error {
	query := `
		UPDATE document_metadata m
		SET corrected_value = $1, corrected_by = $2, corrected_at = NOW()
		FROM documents d
		WHERE m.document_id = d.id AND m.document_id = $3 AND m.key = $4 AND d.tenant_id = $5
	`
	result, err := r.db.Exec(ctx, query, correctedValue, correctedBy, documentID, key, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update metadata field: %w", err)
	}
	if result.RowsAffected() == 0 {
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

// UpdateProcessingFields updates the processing stage, error, and AI suggestion fields.
func (r *documentRepository) UpdateProcessingFields(ctx context.Context, tenantID, documentID string, stage *string, errMsg *string, suggestedFolderName *string, suggestedFilename *string, suggestionConfidence *float32, suggestionCreateNew *bool) error {
	query := `
		UPDATE documents
		SET processing_stage = COALESCE($1, processing_stage),
		    processing_error = COALESCE($2, processing_error),
		    suggested_folder_name = COALESCE($3, suggested_folder_name),
		    suggested_filename = COALESCE($4, suggested_filename),
		    suggestion_confidence = COALESCE($5, suggestion_confidence),
		    suggestion_create_new = COALESCE($6, suggestion_create_new)
		WHERE id = $7 AND tenant_id = $8
	`
	_, execErr := r.db.Exec(ctx, query, stage, errMsg, suggestedFolderName, suggestedFilename, suggestionConfidence, suggestionCreateNew, documentID, tenantID)
	if execErr != nil {
		return fmt.Errorf("failed to update processing fields: %w", execErr)
	}
	return nil
}
