package usecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	model "github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
)

type stubDocumentRepository struct {
	createdDocument *model.Document
	createdVersion  *model.DocumentVersion
}

func (s *stubDocumentRepository) Create(_ context.Context, doc *model.Document) error {
	s.createdDocument = doc
	return nil
}

func (s *stubDocumentRepository) GetByID(context.Context, string, string, string) (*model.Document, error) {
	return nil, nil
}

func (s *stubDocumentRepository) List(context.Context, *repository.ListDocumentsQuery) ([]model.Document, *string, error) {
	return nil, nil, nil
}

func (s *stubDocumentRepository) Update(context.Context, *model.Document) error {
	return nil
}

func (s *stubDocumentRepository) Delete(context.Context, string, string, string, string) error {
	return nil
}

func (s *stubDocumentRepository) CreateVersion(_ context.Context, version *model.DocumentVersion) error {
	s.createdVersion = version
	return nil
}

func (s *stubDocumentRepository) GetVersions(context.Context, string, string) ([]model.DocumentVersion, error) {
	return nil, nil
}

func (s *stubDocumentRepository) CreatePage(context.Context, *model.DocumentPage) error {
	return nil
}

func (s *stubDocumentRepository) GetPages(context.Context, string, string) ([]model.DocumentPage, error) {
	return nil, nil
}

func (s *stubDocumentRepository) SetMetadata(context.Context, string, *model.DocumentMetadata) error {
	return nil
}

func (s *stubDocumentRepository) GetMetadata(context.Context, string, string) ([]model.DocumentMetadata, error) {
	return nil, nil
}

func (s *stubDocumentRepository) UpdateMetadataField(context.Context, string, string, string, string, string) error {
	return nil
}

func (s *stubDocumentRepository) GetFullDocument(context.Context, string, string, string) (*model.Document, []model.DocumentVersion, []model.DocumentMetadata, error) {
	return nil, nil, nil, nil
}

func (s *stubDocumentRepository) UpdateProcessingFields(context.Context, string, string, *string, *string) error {
	return nil
}

func (s *stubDocumentRepository) ClearSuggestion(context.Context, string, string, string, *string) error {
	return nil
}

func (s *stubDocumentRepository) ApplySuggestion(context.Context, *model.Document, *string) error {
	return nil
}

type stubOCRDispatcher struct {
	job *usecase.OCRJob
}

func (s *stubOCRDispatcher) DispatchOCR(_ context.Context, job usecase.OCRJob) error {
	s.job = &job
	return nil
}

func TestUploadPublishesOCRJobContract(t *testing.T) {
	repo := &stubDocumentRepository{}
	dispatcher := &stubOCRDispatcher{}
	svc := usecase.NewDocumentService(repo, nil, nil, dispatcher)

	output, err := svc.Upload(context.Background(), &usecase.UploadDocumentInput{
		TenantID: "tenant-1",
		OrgID:    "org-1",
		OwnerID:  "user-1",
		Title:    "Passport",
		DocType:  "identity",
		File:     newTestFileHeader(t, "passport.pdf", []byte("fake-pdf")),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if output == nil {
		t.Fatal("Upload() returned nil output")
	}

	if dispatcher.job == nil {
		t.Fatal("expected upload to publish an OCR job")
	}

	jobBody, err := json.Marshal(dispatcher.job)
	if err != nil {
		t.Fatalf("failed to marshal dispatched job: %v", err)
	}

	var job map[string]any
	if err := json.Unmarshal(jobBody, &job); err != nil {
		t.Fatalf("failed to unmarshal published job: %v", err)
	}

	for _, field := range []string{"document_id", "version_id", "storage_key", "mime_type", "tenant_id", "org_id"} {
		if _, ok := job[field]; !ok {
			t.Fatalf("published job missing %q: %v", field, job)
		}
	}

	if got := job["retry_count"]; got != float64(0) {
		t.Fatalf("retry_count = %v, want 0", got)
	}
	if _, hasPages := job["pages"]; hasPages {
		t.Fatalf("OCR job should not include pages: %v", job)
	}
	if output.Message != "Document uploaded successfully. OCR will begin shortly." {
		t.Fatalf("unexpected upload message: %q", output.Message)
	}
	if repo.createdDocument == nil || repo.createdVersion == nil {
		t.Fatal("expected document and version records to be created before publishing")
	}
}

func newTestFileHeader(t *testing.T, filename string, contents []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}

	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one uploaded file, got %d", len(files))
	}

	return files[0]
}

var _ repository.DocumentRepository = (*stubDocumentRepository)(nil)
var _ usecase.OCRDispatcher = (*stubOCRDispatcher)(nil)

func (m *stubDocumentRepository) GetStats(ctx context.Context, tenantID, orgID string) (*model.DocumentStats, error) {
	return &model.DocumentStats{}, nil
}
