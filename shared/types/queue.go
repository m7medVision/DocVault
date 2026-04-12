package types

// OCRJob represents a job published to the OCR queue.
type OCRJob struct {
	DocumentID string `json:"document_id"`
	UserID     string `json:"user_id"`
	FileKey    string `json:"file_key"`
	Language   string `json:"language"` // "en", "ar", or "mixed"
	TenantID   string `json:"tenant_id"`
}

// ProcessingJob represents a job published after OCR completes.
type ProcessingJob struct {
	DocumentID string `json:"document_id"`
	UserID     string `json:"user_id"`
	Language   string `json:"language"`
	TenantID   string `json:"tenant_id"`
	TotalPages int    `json:"total_pages"`
}

// ReminderJob represents a job for reminder extraction.
type ReminderJob struct {
	DocumentID    string `json:"document_id"`
	UserID        string `json:"user_id"`
	ExtractedText string `json:"extracted_text"`
	Language      string `json:"language"`
	TenantID      string `json:"tenant_id"`
}
