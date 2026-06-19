package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	auditapp "github.com/docvault/backend/internal/audit/app"
	documentapp "github.com/docvault/backend/internal/document/app"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
)

func (h *Handler) MoveDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.requireDocWritable(w, r, documentID) {
		return
	}

	var body struct {
		FolderID *string `json:"folder_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	input := &documentapp.MoveDocumentInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
		FolderID:   body.FolderID,
	}

	output, err := h.documentSvc.Move(ctx, input)
	if err != nil {
		slog.Error("move document failed", "error", err, "document_id", documentID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"INTERNAL_ERROR"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action": "move",
			"folder": body.FolderID,
		},
	})

	slog.Info("document moved", "document_id", documentID, "folder_id", body.FolderID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":     output.Message,
		"folder_id":   output.FolderID,
		"folder_name": output.FolderName,
	})
}

func (h *Handler) UpdateDocumentTitle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.requireDocWritable(w, r, documentID) {
		return
	}

	var body struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.Title == "" {
		http.Error(w, `{"error":"title is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	input := &documentapp.UpdateTitleInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
		Title:      body.Title,
	}

	if err := h.documentSvc.UpdateTitle(ctx, input); err != nil {
		if errors.Is(err, repository.ErrDocumentTitleExists) {
			http.Error(w, `{"error":"a document with this title already exists in this folder","code":"CONFLICT"}`, http.StatusConflict)
			return
		}
		slog.Error("update title failed", "error", err, "document_id", documentID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"INTERNAL_ERROR"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action": "rename",
			"title":  body.Title,
		},
	})

	slog.Info("document title updated", "document_id", documentID, "title", body.Title, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    documentID,
		"title": body.Title,
	})
}

// AcceptSuggestion applies the document's pending folder suggestion: it
// resolves the suggested nested path (find-or-creating folders), moves the
// document into the resulting leaf folder, optionally retitles it to the
// suggested filename, clears the suggestion columns, and marks processing
// "completed".
func (h *Handler) AcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.requireDocWritable(w, r, documentID) {
		return
	}

	result, err := h.suggestionSvc.Accept(ctx, documentapp.AcceptSuggestionRequest{
		TenantID:   tenantID,
		OrgID:      orgID,
		UserID:     userID,
		DocumentID: documentID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":      "accept_suggestion",
			"folder_id":   result.LeafFolderID,
			"folder_path": result.FolderPath,
			"title":       result.Title,
		},
	})

	slog.Info("suggestion accepted", "document_id", documentID, "folder_id", result.LeafFolderID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document": result.Document,
	})
}

// DismissSuggestion clears a document's pending folder suggestion without moving
// or retitling it.
func (h *Handler) DismissSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.requireDocWritable(w, r, documentID) {
		return
	}

	if err := h.documentSvc.DismissSuggestion(ctx, tenantID, orgID, documentID); err != nil {
		slog.Error("dismiss suggestion failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to dismiss suggestion","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action": "dismiss_suggestion",
		},
	})

	slog.Info("suggestion dismissed", "document_id", documentID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      documentID,
		"message": "suggestion dismissed",
	})
}
