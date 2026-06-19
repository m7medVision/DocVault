// Package usecase contains application business logic.
// Use cases orchestrate operations between transports and infrastructure adapters.
package usecase

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	model "github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

// DocumentService handles document operations.
type DocumentService struct {
	repo          DocumentStore
	aclRepo       DocumentACL
	objectStore   ObjectStore
	ocrDispatcher OCRDispatcher
}

// ObjectStore interface for document storage operations.
type ObjectStore interface {
	PutObject(ctx context.Context, object string, reader io.Reader, objectSize int64, contentType string) error
	PresignGetObject(ctx context.Context, object string, expiry time.Duration) (string, error)
}

// OCRDispatcher dispatches uploaded documents to the OCR pipeline.
type OCRDispatcher interface {
	DispatchOCR(ctx context.Context, job OCRJob) error
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(repo DocumentStore, aclRepo DocumentACL, objectStore ObjectStore, ocrDispatcher OCRDispatcher) *DocumentService {
	return &DocumentService{
		repo:          repo,
		aclRepo:       aclRepo,
		objectStore:   objectStore,
		ocrDispatcher: ocrDispatcher,
	}
}

// UploadDocumentInput contains the data needed to upload a document.
type UploadDocumentInput struct {
	TenantID string
	OrgID    string
	OwnerID  string
	Title    string
	DocType  string
	File     *multipart.FileHeader
	FolderID *string
	Language *string
}

type OCRJob struct {
	DocumentID string  `json:"document_id"`
	VersionID  string  `json:"version_id"`
	StorageKey string  `json:"storage_key"`
	MimeType   string  `json:"mime_type"`
	TenantID   string  `json:"tenant_id"`
	OrgID      string  `json:"org_id"`
	Language   *string `json:"language,omitempty"`
	RetryCount int     `json:"retry_count"`
}

// UploadDocumentOutput contains the result of a document upload.
type UploadDocumentOutput struct {
	DocumentID string
	Message    string
	Status     model.DocumentStatus
}

// ListDocumentsInput contains filters for listing documents.
type ListDocumentsInput struct {
	TenantID string
	OrgID    string
	UserID   string
	GroupIDs []string
	IsAdmin  bool
	DocType  string
	FolderID string
	Status   model.DocumentStatus
	Language string
	Cursor   string
	Limit    int
}

// ListDocumentsOutput contains the result of listing documents.
type ListDocumentsOutput struct {
	Documents []model.Document
	Cursor    *string
	Total     int
}

// GetDocumentInput contains the ID of the document to retrieve.
type GetDocumentInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
}

// GetDocumentOutput contains the full document details.
type GetDocumentOutput struct {
	Document model.Document
	Versions []model.DocumentVersion
	Metadata []model.DocumentMetadata
}

// DeleteDocumentInput contains the ID of the document to delete.
type DeleteDocumentInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
	ActorID    string
}

// DeleteDocumentOutput contains the result of deleting a document.
type DeleteDocumentOutput struct {
	Message string
}

// DownloadDocumentInput contains the ID of the document to download.
type DownloadDocumentInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
	ActorID    string
}

// DownloadDocumentOutput contains the presigned URL for download.
type DownloadDocumentOutput struct {
	PresignedURL string
	ExpiresAt    time.Time
	StorageKey   string
}

// Upload uploads a document, stores it, and queues it for OCR.
func (s *DocumentService) Upload(ctx context.Context, input *UploadDocumentInput) (*UploadDocumentOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner_id is required")
	}
	if input.File == nil {
		return nil, fmt.Errorf("file is required")
	}

	if input.File.Size > model.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum of 50MB")
	}

	docID := uuid.New().String()
	versionID := uuid.New().String()

	// Construct (and validate) the document aggregate before touching storage,
	// so an invalid doc-type fails fast without leaving an orphaned object.
	doc, err := model.NewDocument(model.NewDocumentParams{
		ID:       docID,
		TenantID: input.TenantID,
		OrgID:    input.OrgID,
		OwnerID:  input.OwnerID,
		Title:    input.Title,
		DocType:  input.DocType,
		FolderID: input.FolderID,
		Language: input.Language,
	})
	if err != nil {
		return nil, err
	}

	storageKey := fmt.Sprintf("%s/%s/%s/%s/%s",
		input.TenantID, input.OrgID, docID, versionID, input.File.Filename)

	src, err := input.File.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	if s.objectStore != nil {
		mimeType := input.File.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		if err := s.objectStore.PutObject(ctx, storageKey, src, input.File.Size, mimeType); err != nil {
			return nil, fmt.Errorf("failed to store document: %w", err)
		}
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	version := &model.DocumentVersion{
		ID:            versionID,
		DocumentID:    docID,
		VersionNumber: 1,
		StorageKey:    storageKey,
		MimeType:      input.File.Header.Get("Content-Type"),
		FileSize:      input.File.Size,
		UploadedBy:    &input.OwnerID,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create document version: %w", err)
	}

	if s.ocrDispatcher != nil {
		if err := s.ocrDispatcher.DispatchOCR(ctx, OCRJob{
			DocumentID: docID,
			VersionID:  versionID,
			StorageKey: storageKey,
			MimeType:   version.MimeType,
			TenantID:   input.TenantID,
			OrgID:      input.OrgID,
			Language:   input.Language,
			RetryCount: 0,
		}); err != nil {
			return nil, fmt.Errorf("failed to publish OCR job: %w", err)
		}
	}

	return &UploadDocumentOutput{
		DocumentID: docID,
		Message:    "Document uploaded successfully. OCR will begin shortly.",
		Status:     model.DocumentStatusPending,
	}, nil
}

// List returns documents for a tenant/org with optional filtering.
func (s *DocumentService) List(ctx context.Context, input *ListDocumentsInput) (*ListDocumentsOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}

	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}

	docs, cursor, err := s.aclRepo.ListVisibleDocuments(ctx, repository.ListVisibleParams{
		TenantID: input.TenantID,
		OrgID:    input.OrgID,
		UserID:   input.UserID,
		GroupIDs: input.GroupIDs,
		IsAdmin:  input.IsAdmin,
		DocType:  input.DocType,
		FolderID: input.FolderID,
		Status:   string(input.Status),
		Language: input.Language,
		Cursor:   input.Cursor,
		Limit:    input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return &ListDocumentsOutput{
		Documents: docs,
		Cursor:    cursor,
		Total:     len(docs),
	}, nil
}

// Get returns a single document with all its details.
func (s *DocumentService) Get(ctx context.Context, input *GetDocumentInput) (*GetDocumentOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.DocumentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	doc, versions, metadata, err := s.repo.GetFullDocument(ctx, input.TenantID, input.OrgID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	userFacingKeys := map[string]bool{
		"issuer": true, "amount": true, "currency": true,
		"issue_date": true, "due_date": true, "expiry_date": true, "document_number": true,
	}
	filtered := make([]model.DocumentMetadata, 0, len(metadata))
	for _, m := range metadata {
		if userFacingKeys[m.Key] {
			filtered = append(filtered, m)
		}
	}

	return &GetDocumentOutput{
		Document: *doc,
		Versions: versions,
		Metadata: filtered,
	}, nil
}

// Delete soft-deletes a document.
func (s *DocumentService) Delete(ctx context.Context, input *DeleteDocumentInput) (*DeleteDocumentOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.DocumentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	if err := s.repo.Delete(ctx, input.TenantID, input.OrgID, input.DocumentID, input.ActorID); err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	if s.aclRepo != nil {
		if _, err := s.aclRepo.DeleteGrantsForResource(ctx, repository.ResourceRef{
			TenantID:     input.TenantID,
			OrgID:        input.OrgID,
			ResourceType: "document",
			ResourceID:   input.DocumentID,
		}); err != nil {
			return nil, fmt.Errorf("failed to clean up document grants: %w", err)
		}
	}

	return &DeleteDocumentOutput{
		Message: "Document deleted successfully",
	}, nil
}

// GetDownloadURL generates a presigned URL for document download.
func (s *DocumentService) GetDownloadURL(ctx context.Context, input *DownloadDocumentInput) (*DownloadDocumentOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.DocumentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	versions, err := s.repo.GetVersions(ctx, input.TenantID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document versions: %w", err)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for document")
	}

	latestVersion := versions[0]
	expiresAt := time.Now().Add(15 * time.Minute)

	var presignedURL string
	if s.objectStore != nil {
		presignedURL, err = s.objectStore.PresignGetObject(ctx, latestVersion.StorageKey, 15*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to create presigned download url: %w", err)
		}
	}

	return &DownloadDocumentOutput{
		PresignedURL: presignedURL,
		ExpiresAt:    expiresAt,
		StorageKey:   latestVersion.StorageKey,
	}, nil
}

// ListVersions returns all versions of a document.
func (s *DocumentService) ListVersions(ctx context.Context, tenantID, orgID, documentID string) ([]model.DocumentVersion, error) {
	if tenantID == "" || orgID == "" || documentID == "" {
		return nil, fmt.Errorf("tenant_id, org_id, and document_id are required")
	}

	return s.repo.GetVersions(ctx, tenantID, documentID)
}

// UpdateMetadata updates document metadata fields.
func (s *DocumentService) UpdateMetadata(ctx context.Context, tenantID, orgID, documentID, actorID string, updates map[string]string) error {
	if tenantID == "" || orgID == "" || documentID == "" {
		return fmt.Errorf("tenant_id, org_id, and document_id are required")
	}

	validKeys := map[string]bool{
		"issuer": true, "amount": true, "currency": true,
		"issue_date": true, "expiry_date": true, "document_number": true, "language": true,
	}
	for key := range updates {
		if !validKeys[key] {
			return fmt.Errorf("invalid metadata key: %s", key)
		}
	}

	for key, value := range updates {
		if err := s.repo.UpdateMetadataField(ctx, tenantID, documentID, key, value, actorID); err != nil {
			return fmt.Errorf("failed to update metadata field %s: %w", key, err)
		}
	}

	return nil
}

// GetPages returns OCR text per page with confidence scores.
func (s *DocumentService) GetPages(ctx context.Context, tenantID, orgID, documentID string) ([]model.DocumentPage, error) {
	if tenantID == "" || orgID == "" || documentID == "" {
		return nil, fmt.Errorf("tenant_id, org_id, and document_id are required")
	}

	return s.repo.GetPages(ctx, tenantID, documentID)
}

// MoveDocumentInput contains the data needed to move a document.
type MoveDocumentInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
	FolderID   *string
}

// MoveDocumentOutput contains the result of moving a document.
type MoveDocumentOutput struct {
	Message    string
	FolderID   *string
	FolderName string
}

// Move moves a document to a different folder.
func (s *DocumentService) Move(ctx context.Context, input *MoveDocumentInput) (*MoveDocumentOutput, error) {
	if input.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.DocumentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	doc, err := s.repo.GetByID(ctx, input.TenantID, input.OrgID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	doc.FolderID = input.FolderID
	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to move document: %w", err)
	}

	folderName := ""
	if input.FolderID != nil && *input.FolderID != "" {
		folderName = *input.FolderID
	}

	return &MoveDocumentOutput{
		Message:    "Document moved successfully",
		FolderID:   input.FolderID,
		FolderName: folderName,
	}, nil
}

// UpdateTitleInput contains the data needed to update a document title.
type UpdateTitleInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
	Title      string
}

// UpdateTitle updates a document's title.
func (s *DocumentService) UpdateTitle(ctx context.Context, input *UpdateTitleInput) error {
	if input.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if input.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if input.DocumentID == "" {
		return fmt.Errorf("document_id is required")
	}
	if input.Title == "" {
		return fmt.Errorf("title is required")
	}

	doc, err := s.repo.GetByID(ctx, input.TenantID, input.OrgID, input.DocumentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	doc.Title = input.Title
	if err := s.repo.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update title: %w", err)
	}

	return nil
}

// AcceptSuggestionInput contains the data needed to accept a folder suggestion.
// LeafFolderID is the id of the folder the document should be moved into (the
// leaf of the suggested path, already find-or-created by the caller). Title is
// the suggested display title; when empty the document title is left unchanged.
type AcceptSuggestionInput struct {
	TenantID     string
	OrgID        string
	DocumentID   string
	LeafFolderID string
	Title        string
}

// AcceptSuggestionOutput contains the updated document after accepting.
type AcceptSuggestionOutput struct {
	Document model.Document
}

// AcceptSuggestion moves the document into the resolved leaf folder, optionally
// retitles it, clears the four suggestion columns, and sets processing_stage to
// "completed". The folder path resolution (find-or-create) is performed by the
// caller via FolderService.EnsureFolderPath; this method only commits the
// document-side effects.
//
// The move/retitle and the suggestion clear run inside a single transaction
// (repo.ApplySuggestion) so the two can never diverge: previously they were two
// separate writes and a failure of the second left the document filed but with
// a stale, uncleared suggestion. EnsureFolderPath remains idempotent, so the
// folder resolution performed before this call is safe to retry.
func (s *DocumentService) AcceptSuggestion(ctx context.Context, input *AcceptSuggestionInput) (*AcceptSuggestionOutput, error) {
	if input.TenantID == "" || input.OrgID == "" || input.DocumentID == "" {
		return nil, fmt.Errorf("tenant_id, org_id, and document_id are required")
	}
	if input.LeafFolderID == "" {
		return nil, fmt.Errorf("leaf_folder_id is required")
	}

	doc, err := s.repo.GetByID(ctx, input.TenantID, input.OrgID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	doc.AcceptSuggestion(input.LeafFolderID, input.Title)

	completed := string(model.StageCompleted)
	if err := s.repo.ApplySuggestion(ctx, doc, &completed); err != nil {
		return nil, fmt.Errorf("failed to apply suggestion: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, input.TenantID, input.OrgID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload document: %w", err)
	}

	return &AcceptSuggestionOutput{Document: *updated}, nil
}

// DismissSuggestion clears the four suggestion columns without moving or
// retitling the document. Like accept, it advances processing_stage to
// "completed": dismissing is a terminal decision on the suggestion, so the
// document should no longer report itself as "suggesting".
func (s *DocumentService) DismissSuggestion(ctx context.Context, tenantID, orgID, documentID string) error {
	if tenantID == "" || orgID == "" || documentID == "" {
		return fmt.Errorf("tenant_id, org_id, and document_id are required")
	}
	completed := string(model.StageCompleted)
	if err := s.repo.ClearSuggestion(ctx, tenantID, orgID, documentID, &completed); err != nil {
		return fmt.Errorf("failed to dismiss suggestion: %w", err)
	}
	return nil
}

// UpdateProcessingFieldsInput contains processing stage/suggestion update data.
type UpdateProcessingFieldsInput struct {
	TenantID   string
	OrgID      string
	DocumentID string
	Stage      *string
	ErrorMsg   *string
}

// UpdateProcessingFields updates the processing stage, error, and AI suggestion fields.
func (s *DocumentService) UpdateProcessingFields(ctx context.Context, input *UpdateProcessingFieldsInput) error {
	if input.TenantID == "" || input.OrgID == "" || input.DocumentID == "" {
		return fmt.Errorf("tenant_id, org_id, and document_id are required")
	}
	return s.repo.UpdateProcessingFields(ctx, input.TenantID, input.DocumentID, input.Stage, input.ErrorMsg)
}

type ProcessingStatus struct {
	Status   string
	Message  string
	Progress int
}

func (s *DocumentService) GetProcessingStatus(ctx context.Context, tenantID, orgID, documentID string) (string, string, int) {
	doc, err := s.repo.GetByID(ctx, tenantID, orgID, documentID)
	if err != nil {
		return "unknown", "Document not found", 0
	}

	if doc.ProcessingStage != nil {
		stage := model.ProcessingStage(*doc.ProcessingStage)
		if message, percent, known := stage.Progress(); known {
			if stage.IsFailed() {
				errMsg := ""
				if doc.ProcessingError != nil {
					errMsg = *doc.ProcessingError
				}
				return "failed", errMsg, 0
			}
			return string(stage), message, percent
		}
	}

	switch doc.Status {
	case model.DocumentStatusPending:
		return "pending", "Waiting in queue...", 10
	case model.DocumentStatusProcessing:
		return "processing", "Processing document...", 50
	case model.DocumentStatusProcessed:
		return "completed", "Processing complete", 100
	case model.DocumentStatusFailed:
		return "failed", "Processing failed", 0
	default:
		return string(doc.Status), "", 0
	}
}

func (s *DocumentService) GetStats(ctx context.Context, tenantID, orgID string) (*model.DocumentStats, error) {
	return s.repo.GetStats(ctx, tenantID, orgID)
}
