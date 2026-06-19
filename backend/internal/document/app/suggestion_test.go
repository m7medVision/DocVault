package app

import (
	"context"
	"testing"

	model "github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/platform/apperr"
	"github.com/docvault/backend/internal/repository"
)

type fakeSuggestionDocs struct {
	getDoc       model.Document
	getErr       error
	acceptErr    error
	lastAccept   *AcceptSuggestionInput
	acceptedDoc  model.Document
	acceptCalled bool
}

func (f *fakeSuggestionDocs) Get(_ context.Context, _ *GetDocumentInput) (*GetDocumentOutput, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &GetDocumentOutput{Document: f.getDoc}, nil
}

func (f *fakeSuggestionDocs) AcceptSuggestion(_ context.Context, in *AcceptSuggestionInput) (*AcceptSuggestionOutput, error) {
	f.acceptCalled = true
	f.lastAccept = in
	if f.acceptErr != nil {
		return nil, f.acceptErr
	}
	return &AcceptSuggestionOutput{Document: f.acceptedDoc}, nil
}

type fakeSuggestionFolders struct {
	leafID       string
	err          error
	lastSegments []string
}

func (f *fakeSuggestionFolders) EnsureFolderPath(_ context.Context, _, _, _ string, segments []string) (string, error) {
	f.lastSegments = segments
	if f.err != nil {
		return "", f.err
	}
	return f.leafID, nil
}

func strptr(s string) *string { return &s }

func TestSuggestionAccept_ResolvesPathFilesAndRetitles(t *testing.T) {
	docs := &fakeSuggestionDocs{
		getDoc:      model.Document{SuggestedFolderName: strptr("Finance/Invoices"), SuggestedFilename: strptr("ACME Invoice")},
		acceptedDoc: model.Document{ID: "doc-1", Title: "ACME Invoice"},
	}
	folders := &fakeSuggestionFolders{leafID: "folder-leaf"}
	svc := NewSuggestionService(docs, folders)

	res, err := svc.Accept(context.Background(), AcceptSuggestionRequest{
		TenantID: "t1", OrgID: "o1", UserID: "u1", DocumentID: "doc-1",
	})
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if got, want := folders.lastSegments, []string{"Finance", "Invoices"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("segments = %v, want %v", got, want)
	}
	if !docs.acceptCalled || docs.lastAccept.LeafFolderID != "folder-leaf" {
		t.Fatalf("AcceptSuggestion not called with the resolved leaf folder: %+v", docs.lastAccept)
	}
	if docs.lastAccept.Title != "ACME Invoice" {
		t.Fatalf("title = %q, want %q", docs.lastAccept.Title, "ACME Invoice")
	}
	if res.LeafFolderID != "folder-leaf" || res.FolderPath != "Finance/Invoices" || res.Title != "ACME Invoice" {
		t.Fatalf("result = %+v", res)
	}
}

func TestSuggestionAccept_NoSuggestionIsInvalid(t *testing.T) {
	svc := NewSuggestionService(&fakeSuggestionDocs{getDoc: model.Document{}}, &fakeSuggestionFolders{})
	_, err := svc.Accept(context.Background(), AcceptSuggestionRequest{TenantID: "t1", OrgID: "o1", DocumentID: "d1"})
	if apperr.KindOf(err) != apperr.KindInvalid {
		t.Fatalf("kind = %v, want Invalid (err=%v)", apperr.KindOf(err), err)
	}
}

func TestSuggestionAccept_MissingDocumentIsNotFound(t *testing.T) {
	svc := NewSuggestionService(&fakeSuggestionDocs{getErr: repository.ErrDocumentNotFound}, &fakeSuggestionFolders{})
	_, err := svc.Accept(context.Background(), AcceptSuggestionRequest{TenantID: "t1", OrgID: "o1", DocumentID: "d1"})
	if apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("kind = %v, want NotFound (err=%v)", apperr.KindOf(err), err)
	}
}

func TestSuggestionAccept_TitleCollisionIsConflict(t *testing.T) {
	docs := &fakeSuggestionDocs{
		getDoc:    model.Document{SuggestedFolderName: strptr("Finance"), SuggestedFilename: strptr("dup")},
		acceptErr: repository.ErrDocumentTitleExists,
	}
	svc := NewSuggestionService(docs, &fakeSuggestionFolders{leafID: "leaf"})
	_, err := svc.Accept(context.Background(), AcceptSuggestionRequest{TenantID: "t1", OrgID: "o1", DocumentID: "d1"})
	if apperr.KindOf(err) != apperr.KindConflict {
		t.Fatalf("kind = %v, want Conflict (err=%v)", apperr.KindOf(err), err)
	}
}
