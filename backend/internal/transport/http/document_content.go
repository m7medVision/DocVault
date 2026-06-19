package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	auditapp "github.com/docvault/backend/internal/audit/app"
	documentapp "github.com/docvault/backend/internal/document/app"
	"github.com/docvault/backend/internal/middleware"
)

func (h *Handler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
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

	if h.requireDocVisible(w, r, documentID) {
		return
	}

	input := &documentapp.DownloadDocumentInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
		ActorID:    userID,
	}

	output, err := h.documentSvc.GetDownloadURL(ctx, input)
	if err != nil {
		slog.Error("download document failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to generate download URL","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionDownload,
		Metadata:   nil,
	})

	slog.Info("document download initiated", "document_id", documentID, "tenant_id", tenantID, "org_id", orgID, "user_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"download_url": output.PresignedURL,
		"expires_at":   output.ExpiresAt,
		"storage_key":  output.StorageKey,
	})
}

func (h *Handler) ListDocumentVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
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

	if h.requireDocVisible(w, r, documentID) {
		return
	}

	versions, err := h.documentSvc.ListVersions(ctx, tenantID, orgID, documentID)
	if err != nil {
		slog.Error("list versions failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to list versions","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document_id": documentID,
		"versions":    versions,
	})
}

func (h *Handler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
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

	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if len(updates) == 0 {
		http.Error(w, `{"error":"no updates provided","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := h.documentSvc.UpdateMetadata(ctx, tenantID, orgID, documentID, userID, updates); err != nil {
		slog.Error("update metadata failed", "error", err, "document_id", documentID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	updatedFields := make([]string, 0, len(updates))
	for k := range updates {
		updatedFields = append(updatedFields, k)
	}
	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"fields": updatedFields,
		},
	})

	slog.Info("metadata updated", "document_id", documentID, "tenant_id", tenantID, "actor_id", userID, "fields", len(updates))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      documentID,
		"message": "metadata updated successfully",
		"updated": updates,
	})
}

func (h *Handler) GetDocumentPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
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

	if h.requireDocVisible(w, r, documentID) {
		return
	}

	pages, err := h.documentSvc.GetPages(ctx, tenantID, orgID, documentID)
	if err != nil {
		slog.Error("get pages failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to get pages","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document_id": documentID,
		"pages":       pages,
	})
}
