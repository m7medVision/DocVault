package app

import (
	"context"
	"errors"
	"fmt"

	model "github.com/docvault/backend/internal/document"
	"github.com/docvault/backend/internal/platform/apperr"
	"github.com/docvault/backend/internal/repository"
)

// SuggestionDocuments is the document-side surface the SuggestionService needs:
// load a document and commit an accepted folder suggestion. *DocumentService
// satisfies it.
type SuggestionDocuments interface {
	Get(ctx context.Context, input *GetDocumentInput) (*GetDocumentOutput, error)
	AcceptSuggestion(ctx context.Context, input *AcceptSuggestionInput) (*AcceptSuggestionOutput, error)
}

// SuggestionFolders is the folder-side surface the SuggestionService needs:
// find-or-create the suggested folder path. *FolderService satisfies it.
type SuggestionFolders interface {
	EnsureFolderPath(ctx context.Context, tenantID, orgID, userID string, segments []string) (string, error)
}

// SuggestionService orchestrates accepting an AI folder suggestion, a flow that
// spans two aggregates: it resolves (find-or-creates) the suggested folder path
// via the folder aggregate, then files and retitles the document via the
// document aggregate. This cross-aggregate orchestration previously lived in the
// HTTP handler; keeping it in the app layer leaves the transport thin and makes
// the flow unit-testable.
type SuggestionService struct {
	docs    SuggestionDocuments
	folders SuggestionFolders
}

// NewSuggestionService wires the orchestration over the document and folder
// services.
func NewSuggestionService(docs SuggestionDocuments, folders SuggestionFolders) *SuggestionService {
	return &SuggestionService{docs: docs, folders: folders}
}

// AcceptSuggestionRequest identifies the document whose pending suggestion is
// being accepted and the principal performing the action.
type AcceptSuggestionRequest struct {
	TenantID   string
	OrgID      string
	UserID     string
	DocumentID string
}

// AcceptSuggestionResult is the accepted document plus the resolved folder
// details the caller needs for its audit trail.
type AcceptSuggestionResult struct {
	Document     model.Document
	LeafFolderID string
	FolderPath   string
	Title        string
}

// Accept resolves the document's suggested folder path, files the document into
// the resolved leaf folder, applies the suggested title, and clears the
// suggestion. Failure modes are returned as typed apperr values so the transport
// renders them uniformly: a missing suggestion or empty path is Invalid, an
// absent document is NotFound, and a title collision is Conflict.
func (s *SuggestionService) Accept(ctx context.Context, req AcceptSuggestionRequest) (*AcceptSuggestionResult, error) {
	getOutput, err := s.docs.Get(ctx, &GetDocumentInput{
		TenantID:   req.TenantID,
		OrgID:      req.OrgID,
		DocumentID: req.DocumentID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return nil, apperr.NewNotFound("NOT_FOUND", "document not found")
		}
		return nil, fmt.Errorf("failed to load document: %w", err)
	}

	doc := getOutput.Document
	if doc.SuggestedFolderName == nil || *doc.SuggestedFolderName == "" {
		return nil, apperr.NewInvalid("BAD_REQUEST", "no folder suggestion to accept")
	}

	segments := SplitFolderPath(*doc.SuggestedFolderName)
	if len(segments) == 0 {
		return nil, apperr.NewInvalid("BAD_REQUEST", "suggested folder path is empty")
	}

	leafID, err := s.folders.EnsureFolderPath(ctx, req.TenantID, req.OrgID, req.UserID, segments)
	if err != nil {
		return nil, fmt.Errorf("failed to create suggested folders: %w", err)
	}

	title := ""
	if doc.SuggestedFilename != nil {
		title = *doc.SuggestedFilename
	}

	output, err := s.docs.AcceptSuggestion(ctx, &AcceptSuggestionInput{
		TenantID:     req.TenantID,
		OrgID:        req.OrgID,
		DocumentID:   req.DocumentID,
		LeafFolderID: leafID,
		Title:        title,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDocumentTitleExists) {
			return nil, apperr.NewConflict("CONFLICT", "a document with this title already exists in this folder")
		}
		return nil, fmt.Errorf("failed to accept suggestion: %w", err)
	}

	return &AcceptSuggestionResult{
		Document:     output.Document,
		LeafFolderID: leafID,
		FolderPath:   *doc.SuggestedFolderName,
		Title:        title,
	}, nil
}
