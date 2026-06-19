package document

import (
	"errors"
	"time"
)

// ErrMissingScope is returned when a document is constructed without the
// tenant/org/owner identifiers every document must carry.
var ErrMissingScope = errors.New("tenant_id, org_id, and owner_id are required")

// NewDocumentParams carries the inputs for constructing a Document aggregate.
type NewDocumentParams struct {
	ID       string
	TenantID string
	OrgID    string
	OwnerID  string
	Title    string
	DocType  string
	FolderID *string
	Language *string
}

// NewDocument builds a Document in its initial Pending state, enforcing the
// invariants every document must satisfy: a complete tenant/org/owner scope and
// a recognised document type. It returns ErrMissingScope or ErrInvalidDocType
// (wrapped) when those are not met.
func NewDocument(p NewDocumentParams) (*Document, error) {
	if p.TenantID == "" || p.OrgID == "" || p.OwnerID == "" {
		return nil, ErrMissingScope
	}
	docType, err := ParseDocType(p.DocType)
	if err != nil {
		return nil, err
	}
	return &Document{
		ID:        p.ID,
		TenantID:  p.TenantID,
		OrgID:     p.OrgID,
		FolderID:  p.FolderID,
		OwnerID:   p.OwnerID,
		Title:     p.Title,
		DocType:   string(docType),
		Status:    DocumentStatusPending,
		Language:  p.Language,
		CreatedAt: time.Now(),
	}, nil
}

// AcceptSuggestion files the document into the resolved leaf folder and, when a
// non-empty title is supplied, applies the suggested title. Clearing the
// suggestion columns and advancing the processing stage are persistence
// concerns handled atomically by the repository.
func (d *Document) AcceptSuggestion(leafFolderID, title string) {
	folder := leafFolderID
	d.FolderID = &folder
	if title != "" {
		d.Title = title
	}
}
