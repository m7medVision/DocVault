package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	auditapp "github.com/docvault/backend/internal/audit/app"
	"github.com/docvault/backend/internal/document"
	documentapp "github.com/docvault/backend/internal/document/app"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
)

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions for document upload","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse multipart form: %s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}

	docType := r.FormValue("doc_type")
	if docType == "" {
		docType = "other"
	}

	folderID := r.FormValue("folder_id")
	if folderID == "" {
		folderID = r.FormValue("folderId")
	}

	language := r.FormValue("language")
	if language == "" {
		language = r.FormValue("lang")
	}

	var folderIDPtr *string
	if folderID != "" {
		folderIDPtr = &folderID
	}

	var languagePtr *string
	if language != "" {
		languagePtr = &language
	}

	input := &documentapp.UploadDocumentInput{
		TenantID: tenantID,
		OrgID:    orgID,
		OwnerID:  userID,
		Title:    title,
		DocType:  docType,
		FolderID: folderIDPtr,
		Language: languagePtr,
		File:     header,
	}
	output, err := h.documentSvc.Upload(ctx, input)
	if err != nil {
		slog.Error("document upload failed", "error", err, "tenant_id", tenantID, "org_id", orgID)
		http.Error(w, fmt.Sprintf(`{"error":"upload failed: %s","code":"INTERNAL_ERROR"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	slog.Info("document uploaded", "document_id", output.DocumentID, "tenant_id", tenantID, "org_id", orgID, "user_id", userID)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"id":       output.DocumentID,
		"status":   output.Status,
		"message":  output.Message,
		"title":    title,
		"doc_type": docType,
	})
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
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

	isAdmin := middleware.HasMinRole(role, middleware.RoleAdmin)
	groupIDs, err := h.aclRepo.ListUserGroupIDs(ctx, userID, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to resolve permissions","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	docType := r.URL.Query().Get("type")
	folderID := r.URL.Query().Get("folder_id")
	status := r.URL.Query().Get("status")
	language := r.URL.Query().Get("language")
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	input := &documentapp.ListDocumentsInput{
		TenantID: tenantID,
		OrgID:    orgID,
		UserID:   userID,
		GroupIDs: groupIDs,
		IsAdmin:  isAdmin,
		DocType:  docType,
		FolderID: folderID,
		Status:   document.DocumentStatus(status),
		Language: language,
		Cursor:   cursor,
		Limit:    limit,
	}

	output, err := h.documentSvc.List(ctx, input)
	if err != nil {
		slog.Error("list documents failed", "error", err)
		http.Error(w, `{"error":"failed to list documents","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"documents": output.Documents,
		"cursor":    output.Cursor,
		"total":     output.Total,
	})
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
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

	input := &documentapp.GetDocumentInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
	}

	output, err := h.documentSvc.Get(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			http.Error(w, `{"error":"document not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("get document failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to get document","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document": output.Document,
		"versions": output.Versions,
		"metadata": output.Metadata,
	})
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanDelete(role) {
		http.Error(w, `{"error":"insufficient permissions for deletion","code":"FORBIDDEN"}`, http.StatusForbidden)
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

	input := &documentapp.DeleteDocumentInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
		ActorID:    userID,
	}

	output, err := h.documentSvc.Delete(ctx, input)
	if err != nil {
		slog.Error("delete document failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to delete document","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     auditapp.AuditActionDelete,
		Metadata:   nil,
	})

	slog.Info("document deleted", "document_id", documentID, "tenant_id", tenantID, "org_id", orgID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": output.Message,
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	stats, err := h.documentSvc.GetStats(ctx, tenantID, orgID)
	if err != nil {
		slog.Error("failed to get stats", "error", err, "tenant_id", tenantID, "org_id", orgID)
		respondError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
