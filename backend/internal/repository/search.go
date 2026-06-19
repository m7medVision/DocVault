package repository

import (
	"context"
	"time"
)

// DocumentChunk is an embeddable unit of document text.
type DocumentChunk struct {
	ID         string
	DocumentID string
	ChunkText  string
	PageNumber int
	Embedding  []float32
	TenantID   string
}

// ChunkMatch is a single retrieval hit returned by the search query.
type ChunkMatch struct {
	DocumentID    string
	DocumentTitle string
	DocType       string
	ChunkID       string
	ChunkText     string
	PageNumber    int
	Language      string
	IsTranslation bool
	Score         float64
}

// DocFact is a single extracted (or user-corrected) key/value fact for a
// document, e.g. issuer, amount, issue/expiry date, document number. Used to
// ground chat answers in the structured data the classifier already pulled.
type DocFact struct {
	DocumentID string
	Key        string
	Value      string
}

// SearchRequest is the filter/identity bundle a retrieval call carries. The
// UserID/GroupIDs/IsAdmin fields drive the visibility seam.
type SearchRequest struct {
	Query       string
	TenantID    string
	UserID      string
	GroupIDs    []string
	IsAdmin     bool
	OrgID       string
	DocumentID  string
	DocType     string
	Language    string
	Status      string
	FolderID    string
	Tags        []string
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Cursor      string
	QueryVector string
	MinScore    float64
}

// SearchResult is a page of retrieval hits.
type SearchResult struct {
	Chunks     []ChunkMatch
	NextCursor string
	TotalCount int
}

// SearchRepository provides vector/full-text retrieval over document chunks,
// gated by the visibility seam.
type SearchRepository interface {
	Search(ctx context.Context, req SearchRequest) (*SearchResult, error)
	IndexChunk(ctx context.Context, chunk DocumentChunk) error
	DeleteChunksByDocument(ctx context.Context, docID string) error
	// FetchDocumentsMetadata returns the extracted (or corrected) facts for the
	// given documents, grouped by document id. Tenant-scoped like every read.
	FetchDocumentsMetadata(ctx context.Context, tenantID string, docIDs []string) (map[string][]DocFact, error)
}
