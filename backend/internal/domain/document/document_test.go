package document

import (
	"errors"
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc, err := NewDocument(NewDocumentParams{
		ID:       "doc-1",
		TenantID: "t1",
		OrgID:    "o1",
		OwnerID:  "u1",
		Title:    "Invoice",
		DocType:  "invoice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Status != DocumentStatusPending {
		t.Errorf("status = %q, want pending", doc.Status)
	}
	if doc.DocType != "invoice" || doc.ID != "doc-1" {
		t.Errorf("unexpected document: %+v", doc)
	}
	if doc.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewDocumentInvariants(t *testing.T) {
	if _, err := NewDocument(NewDocumentParams{OrgID: "o1", OwnerID: "u1", DocType: "invoice"}); !errors.Is(err, ErrMissingScope) {
		t.Errorf("missing tenant: err = %v, want ErrMissingScope", err)
	}
	if _, err := NewDocument(NewDocumentParams{TenantID: "t1", OrgID: "o1", OwnerID: "u1", DocType: "passport"}); !errors.Is(err, ErrInvalidDocType) {
		t.Errorf("bad doctype: err = %v, want ErrInvalidDocType", err)
	}
}

func TestAcceptSuggestion(t *testing.T) {
	d := &Document{Title: "old"}
	d.AcceptSuggestion("folder-9", "New Title")
	if d.FolderID == nil || *d.FolderID != "folder-9" {
		t.Errorf("FolderID = %v, want folder-9", d.FolderID)
	}
	if d.Title != "New Title" {
		t.Errorf("Title = %q, want New Title", d.Title)
	}

	// An empty title leaves the existing title untouched.
	d2 := &Document{Title: "keep"}
	d2.AcceptSuggestion("folder-1", "")
	if d2.Title != "keep" {
		t.Errorf("Title = %q, want unchanged", d2.Title)
	}
}
