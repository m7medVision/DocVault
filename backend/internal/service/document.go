// Package service contains business logic for the application.
// Services orchestrate operations between handlers and repositories.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

// DocumentService handles document operations.
type DocumentService struct {
	repo        repository.DocumentRepository
	objectStore ObjectStore
	publisher   QueuePublisher
}

// ObjectStore interface for document storage operations.
type ObjectStore interface {
	PutObject(ctx context.Context, object string, reader io.Reader, objectSize int64, contentType string) error
	PresignGetObject(ctx context.Context, object string, expiry time.Duration) (string, error)
}

// QueuePublisher interface for publishing messages to RabbitMQ.
type QueuePublisher interface {
	Publish(ctx context.Context, body []byte) error
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(repo repository.DocumentRepository, objectStore ObjectStore, publisher QueuePublisher) *DocumentService {
	return &DocumentService{
		repo:        repo,
		objectStore: objectStore,
		publisher:   publisher,
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

type processingJob struct {
	DocumentID string  `json:"document_id"`
	VersionID  string  `json:"version_id"`
	StorageKey string  `json:"storage_key"`
	MimeType   string  `json:"mime_type"`
	TenantID   string  `json:"tenant_id"`
	OrgID      string  `json:"org_id"`
	Language   *string `json:"language,omitempty"`
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

// Upload uploads a document, stores it, and queues it for processing.
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

	const maxFileSize = 50 * 1024 * 1024
	if input.File.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum of 50MB")
	}

	validDocTypes := map[string]bool{
		"invoice": true, "contract": true, "identity": true,
		"warranty": true, "receipt": true, "other": true,
	}
	if !validDocTypes[input.DocType] {
		return nil, fmt.Errorf("invalid document type: %s", input.DocType)
	}

	docID := uuid.New().String()
	versionID := uuid.New().String()

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

	doc := &model.Document{
		ID:        docID,
		TenantID:  input.TenantID,
		OrgID:     input.OrgID,
		FolderID:  input.FolderID,
		OwnerID:   input.OwnerID,
		Title:     input.Title,
		DocType:   input.DocType,
		Status:    model.DocumentStatusPending,
		Language:  input.Language,
		CreatedAt: time.Now(),
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

	if s.publisher != nil {
		jobBody, err := json.Marshal(processingJob{
			DocumentID: docID,
			VersionID:  versionID,
			StorageKey: storageKey,
			MimeType:   version.MimeType,
			TenantID:   input.TenantID,
			OrgID:      input.OrgID,
			Language:   input.Language,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processing job: %w", err)
		}
		if err := s.publisher.Publish(ctx, jobBody); err != nil {
			return nil, fmt.Errorf("failed to publish processing job: %w", err)
		}
	}

	return &UploadDocumentOutput{
		DocumentID: docID,
		Message:    "Document uploaded successfully. Processing will begin shortly.",
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

	query := &repository.ListDocumentsQuery{
		TenantID: input.TenantID,
		OrgID:    input.OrgID,
		DocType:  input.DocType,
		FolderID: input.FolderID,
		Status:   input.Status,
		Language: input.Language,
		Cursor:   input.Cursor,
		Limit:    input.Limit,
	}

	docs, cursor, err := s.repo.List(ctx, query)
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

	return &GetDocumentOutput{
		Document: *doc,
		Versions: versions,
		Metadata: metadata,
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
