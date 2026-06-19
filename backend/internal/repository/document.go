package repository

import model "github.com/docvault/backend/internal/document"

// ListDocumentsQuery contains filters for listing documents. It is the contract
// the document service passes to the data-access layer; the postgres adapter
// translates it into the generated sqlc params.
type ListDocumentsQuery struct {
	TenantID string
	OrgID    string
	DocType  string
	FolderID string
	Status   model.DocumentStatus
	Language string
	Cursor   string
	Limit    int
}
