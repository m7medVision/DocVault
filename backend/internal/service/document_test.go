package service_test

import (
	"context"
	"mime/multipart"
	"testing"

	"github.com/docvault/backend/internal/service"
)

// TestDocumentServiceUploadValidation tests validation in document upload.
func TestDocumentServiceUploadValidation(t *testing.T) {
	svc := service.NewDocumentService(nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.UploadDocumentInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.UploadDocumentInput{
				TenantID: "",
				OrgID:    "org-1",
				OwnerID:  "user-1",
				File:     &multipart.FileHeader{Filename: "test.pdf", Size: 1024},
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing org_id",
			input: &service.UploadDocumentInput{
				TenantID: "tenant-1",
				OrgID:    "",
				OwnerID:  "user-1",
				File:     &multipart.FileHeader{Filename: "test.pdf", Size: 1024},
			},
			wantErr: true,
			errMsg:  "org_id is required",
		},
		{
			name: "missing owner_id",
			input: &service.UploadDocumentInput{
				TenantID: "tenant-1",
				OrgID:    "org-1",
				OwnerID:  "",
				File:     &multipart.FileHeader{Filename: "test.pdf", Size: 1024},
			},
			wantErr: true,
			errMsg:  "owner_id is required",
		},
		{
			name: "missing file",
			input: &service.UploadDocumentInput{
				TenantID: "tenant-1",
				OrgID:    "org-1",
				OwnerID:  "user-1",
				File:     nil,
			},
			wantErr: true,
			errMsg:  "file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Upload(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestDocumentServiceListValidation tests validation in document listing.
func TestDocumentServiceListValidation(t *testing.T) {
	svc := service.NewDocumentService(nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.ListDocumentsInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.ListDocumentsInput{
				TenantID: "",
				OrgID:    "org-1",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing org_id",
			input: &service.ListDocumentsInput{
				TenantID: "tenant-1",
				OrgID:    "",
			},
			wantErr: true,
			errMsg:  "org_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.List(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestDocumentServiceDeleteValidation tests validation in document deletion.
func TestDocumentServiceDeleteValidation(t *testing.T) {
	svc := service.NewDocumentService(nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.DeleteDocumentInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.DeleteDocumentInput{
				TenantID:   "",
				OrgID:      "org-1",
				DocumentID: "doc-1",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing org_id",
			input: &service.DeleteDocumentInput{
				TenantID:   "tenant-1",
				OrgID:      "",
				DocumentID: "doc-1",
			},
			wantErr: true,
			errMsg:  "org_id is required",
		},
		{
			name: "missing document_id",
			input: &service.DeleteDocumentInput{
				TenantID:   "tenant-1",
				OrgID:      "org-1",
				DocumentID: "",
			},
			wantErr: true,
			errMsg:  "document_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Delete(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestDocumentServiceGetValidation tests validation in getting document.
func TestDocumentServiceGetValidation(t *testing.T) {
	svc := service.NewDocumentService(nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.GetDocumentInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.GetDocumentInput{
				TenantID:   "",
				OrgID:      "org-1",
				DocumentID: "doc-1",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing document_id",
			input: &service.GetDocumentInput{
				TenantID:   "tenant-1",
				OrgID:      "org-1",
				DocumentID: "",
			},
			wantErr: true,
			errMsg:  "document_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Get(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestFolderServiceCreateValidation tests validation in folder creation.
func TestFolderServiceCreateValidation(t *testing.T) {
	svc := service.NewFolderService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.CreateFolderInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.CreateFolderInput{
				TenantID: "",
				OrgID:    "org-1",
				Name:     "Test Folder",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing org_id",
			input: &service.CreateFolderInput{
				TenantID: "tenant-1",
				OrgID:    "",
				Name:     "Test Folder",
			},
			wantErr: true,
			errMsg:  "org_id is required",
		},
		{
			name: "missing name",
			input: &service.CreateFolderInput{
				TenantID: "tenant-1",
				OrgID:    "org-1",
				Name:     "",
			},
			wantErr: true,
			errMsg:  "folder name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestTagServiceListValidation tests validation in tag listing.
func TestTagServiceListValidation(t *testing.T) {
	svc := service.NewTagService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *service.ListTagsInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing tenant_id",
			input: &service.ListTagsInput{
				TenantID: "",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.List(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
