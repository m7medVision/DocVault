package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
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

	input := &usecase.UploadDocumentInput{
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

	input := &usecase.ListDocumentsInput{
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

	input := &usecase.GetDocumentInput{
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

	input := &usecase.DeleteDocumentInput{
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

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionDelete,
		Metadata:   nil,
	})

	slog.Info("document deleted", "document_id", documentID, "tenant_id", tenantID, "org_id", orgID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": output.Message,
	})
}

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

	input := &usecase.DownloadDocumentInput{
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

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionDownload,
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
	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
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

	input := &usecase.MoveDocumentInput{
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

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
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

	input := &usecase.UpdateTitleInput{
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

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
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

	getOutput, err := h.documentSvc.Get(ctx, &usecase.GetDocumentInput{
		TenantID:   tenantID,
		OrgID:      orgID,
		DocumentID: documentID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			http.Error(w, `{"error":"document not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("accept suggestion: load document failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to load document","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	doc := getOutput.Document
	if doc.SuggestedFolderName == nil || *doc.SuggestedFolderName == "" {
		http.Error(w, `{"error":"no folder suggestion to accept","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	segments := usecase.SplitFolderPath(*doc.SuggestedFolderName)
	if len(segments) == 0 {
		http.Error(w, `{"error":"suggested folder path is empty","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	leafID, err := h.folderSvc.EnsureFolderPath(ctx, tenantID, orgID, userID, segments)
	if err != nil {
		slog.Error("accept suggestion: ensure folder path failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to create suggested folders","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	title := ""
	if doc.SuggestedFilename != nil {
		title = *doc.SuggestedFilename
	}

	output, err := h.documentSvc.AcceptSuggestion(ctx, &usecase.AcceptSuggestionInput{
		TenantID:     tenantID,
		OrgID:        orgID,
		DocumentID:   documentID,
		LeafFolderID: leafID,
		Title:        title,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDocumentTitleExists) {
			http.Error(w, `{"error":"a document with this title already exists in this folder","code":"CONFLICT"}`, http.StatusConflict)
			return
		}
		slog.Error("accept suggestion failed", "error", err, "document_id", documentID)
		http.Error(w, `{"error":"failed to accept suggestion","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":      "accept_suggestion",
			"folder_id":   leafID,
			"folder_path": *doc.SuggestedFolderName,
			"title":       title,
		},
	})

	slog.Info("suggestion accepted", "document_id", documentID, "folder_id", leafID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document": output.Document,
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

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
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
