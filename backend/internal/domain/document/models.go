package document

import "time"

// DocumentStatus represents the coarse processing state of a document.
type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusProcessed  DocumentStatus = "processed"
	DocumentStatusFailed     DocumentStatus = "failed"
)

// ProcessingStage represents the granular pipeline stage of a document.
type ProcessingStage string

const (
	StageUploaded          ProcessingStage = "uploaded"
	StageOCRQueued         ProcessingStage = "ocr_queued"
	StageOCRRunning        ProcessingStage = "ocr_running"
	StageOCRComplete       ProcessingStage = "ocr_complete"
	StageProcessingQueued  ProcessingStage = "processing_queued"
	StageProcessingRunning ProcessingStage = "processing_running"
	StageIndexing          ProcessingStage = "indexing"
	StageSuggesting        ProcessingStage = "suggesting"
	StageCompleted         ProcessingStage = "completed"
	StageOCRFailed         ProcessingStage = "ocr_failed"
	StageProcessingFailed  ProcessingStage = "processing_failed"
)

// Document represents a stable identity record for an uploaded file.
type Document struct {
	ID                   string         `json:"id"`
	TenantID             string         `json:"tenant_id"`
	OrgID                string         `json:"org_id"`
	FolderID             *string        `json:"folder_id,omitempty"`
	OwnerID              string         `json:"owner_id"`
	Title                string         `json:"title"`
	DocType              string         `json:"doc_type"`
	Status               DocumentStatus `json:"status"`
	Language             *string        `json:"language,omitempty"`
	ProcessingStage      *string        `json:"processing_stage,omitempty"`
	ProcessingError      *string        `json:"processing_error,omitempty"`
	SuggestedFolderName  *string        `json:"suggested_folder_name"`
	SuggestedFilename    *string        `json:"suggested_filename"`
	SuggestionConfidence *float32       `json:"suggestion_confidence"`
	SuggestionCreateNew  *bool          `json:"suggestion_create_new"`
	CreatedAt            time.Time      `json:"created_at"`
}

// DocumentVersion represents a single version of a document.
type DocumentVersion struct {
	ID            string    `json:"id"`
	DocumentID    string    `json:"document_id"`
	VersionNumber int       `json:"version_number"`
	StorageKey    string    `json:"storage_key"`
	MimeType      string    `json:"mime_type"`
	FileSize      int64     `json:"file_size"`
	UploadedBy    *string   `json:"uploaded_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// DocumentPage represents OCR output for a single page.
type DocumentPage struct {
	ID             string    `json:"id"`
	DocumentID     string    `json:"document_id"`
	VersionID      string    `json:"version_id"`
	PageNumber     int       `json:"page_number"`
	OCRText        *string   `json:"ocr_text,omitempty"`
	TranslatedText *string   `json:"translated_text,omitempty"`
	Confidence     *float32  `json:"confidence,omitempty"`
	OCRModel       string    `json:"ocr_model"`
	CreatedAt      time.Time `json:"created_at"`
}

// DocumentMetadata represents key-value metadata for a document.
type DocumentMetadata struct {
	ID             string     `json:"id"`
	DocumentID     string     `json:"document_id"`
	Key            string     `json:"key"`
	ExtractedValue *string    `json:"extracted_value,omitempty"`
	CorrectedValue *string    `json:"corrected_value,omitempty"`
	CorrectedBy    *string    `json:"corrected_by,omitempty"`
	CorrectedAt    *time.Time `json:"corrected_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ExtractedTextChunk represents a retrieval unit with embedding.
type ExtractedTextChunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	PageID     string    `json:"page_id"`
	ChunkIndex int       `json:"chunk_index"`
	ChunkText  string    `json:"chunk_text"`
	CreatedAt  time.Time `json:"created_at"`
}

// Folder represents a nested folder tree node.
type Folder struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	OrgID     string    `json:"org_id"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Tag represents a free-form label.
type Tag struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// DocumentStats represents summary metrics for a tenant's documents.
type DocumentStats struct {
	TotalDocuments    int64 `json:"total_documents"`
	PendingDocuments  int64 `json:"pending_documents"`
	CompletedThisWeek int64 `json:"completed_this_week"`
	StorageUsedBytes  int64 `json:"storage_used_bytes"`
}
