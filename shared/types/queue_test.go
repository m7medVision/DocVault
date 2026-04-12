package types

import (
	"encoding/json"
	"testing"
)

func TestOCRJobJSON(t *testing.T) {
	job := OCRJob{
		DocumentID: "doc-123",
		UserID:     "user-456",
		FileKey:    "files/doc-123.pdf",
		Language:   "en",
		TenantID:   "tenant-789",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal OCRJob: %v", err)
	}

	var unmarshaled OCRJob
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal OCRJob: %v", err)
	}

	if unmarshaled.DocumentID != job.DocumentID {
		t.Errorf("DocumentID: got %q, want %q", unmarshaled.DocumentID, job.DocumentID)
	}
	if unmarshaled.UserID != job.UserID {
		t.Errorf("UserID: got %q, want %q", unmarshaled.UserID, job.UserID)
	}
	if unmarshaled.FileKey != job.FileKey {
		t.Errorf("FileKey: got %q, want %q", unmarshaled.FileKey, job.FileKey)
	}
	if unmarshaled.Language != job.Language {
		t.Errorf("Language: got %q, want %q", unmarshaled.Language, job.Language)
	}
	if unmarshaled.TenantID != job.TenantID {
		t.Errorf("TenantID: got %q, want %q", unmarshaled.TenantID, job.TenantID)
	}
}

func TestProcessingJobJSON(t *testing.T) {
	job := ProcessingJob{
		DocumentID: "doc-123",
		UserID:     "user-456",
		Language:   "ar",
		TenantID:   "tenant-789",
		TotalPages: 10,
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal ProcessingJob: %v", err)
	}

	var unmarshaled ProcessingJob
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ProcessingJob: %v", err)
	}

	if unmarshaled.TotalPages != job.TotalPages {
		t.Errorf("TotalPages: got %d, want %d", unmarshaled.TotalPages, job.TotalPages)
	}
}

func TestReminderJobJSON(t *testing.T) {
	job := ReminderJob{
		DocumentID:    "doc-123",
		UserID:        "user-456",
		ExtractedText: "Meeting tomorrow at 3pm",
		Language:      "en",
		TenantID:      "tenant-789",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal ReminderJob: %v", err)
	}

	var unmarshaled ReminderJob
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ReminderJob: %v", err)
	}

	if unmarshaled.ExtractedText != job.ExtractedText {
		t.Errorf("ExtractedText: got %q, want %q", unmarshaled.ExtractedText, job.ExtractedText)
	}
}
