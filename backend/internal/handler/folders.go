package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/service"
)

func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, `{"error":"folder name is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	input := &service.CreateFolderInput{
		TenantID:  tenantID,
		OrgID:     orgID,
		ParentID:  body.ParentID,
		Name:      body.Name,
		CreatedBy: userID,
	}

	output, err := h.folderSvc.Create(ctx, input)
	if err != nil {
		slog.Error("create folder failed", "error", err, "tenant_id", tenantID, "org_id", orgID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	h.auditSvc.Write(ctx, &service.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   output.Folder.ID,
		Action:     service.AuditActionCreate,
		Metadata:   nil,
	})

	slog.Info("folder created", "folder_id", output.Folder.ID, "tenant_id", tenantID, "org_id", orgID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"folder": output.Folder,
	})
}

func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
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

	parentIDStr := r.URL.Query().Get("parent_id")
	var parentID *string
	if parentIDStr != "" {
		parentID = &parentIDStr
	}

	input := &service.ListFoldersInput{
		TenantID: tenantID,
		OrgID:    orgID,
		ParentID: parentID,
	}

	output, err := h.folderSvc.List(ctx, input)
	if err != nil {
		slog.Error("list folders failed", "error", err, "tenant_id", tenantID, "org_id", orgID)
		http.Error(w, `{"error":"failed to list folders","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"folders": output.Folders,
	})
}

func (h *Handler) ListAllFolders(w http.ResponseWriter, r *http.Request) {
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

	folders, err := h.folderSvc.ListAll(ctx, tenantID, orgID)
	if err != nil {
		slog.Error("list all folders failed", "error", err, "tenant_id", tenantID, "org_id", orgID)
		http.Error(w, `{"error":"failed to list folders","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"folders": folders,
	})
}

func (h *Handler) RenameFolder(w http.ResponseWriter, r *http.Request) {
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

	folderID := r.PathValue("id")
	if folderID == "" {
		http.Error(w, `{"error":"folder id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, `{"error":"folder name is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := h.folderSvc.Rename(ctx, tenantID, orgID, folderID, body.Name); err != nil {
		slog.Error("rename folder failed", "error", err, "folder_id", folderID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	h.auditSvc.Write(ctx, &service.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     service.AuditActionUpdate,
		Metadata:   map[string]interface{}{"action": "rename", "name": body.Name},
	})

	slog.Info("folder renamed", "folder_id", folderID, "name", body.Name, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":   folderID,
		"name": body.Name,
	})
}

func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanDelete(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	folderID := r.PathValue("id")
	if folderID == "" {
		http.Error(w, `{"error":"folder id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := h.folderSvc.Delete(ctx, tenantID, orgID, folderID); err != nil {
		slog.Error("delete folder failed", "error", err, "folder_id", folderID)
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	h.auditSvc.Write(ctx, &service.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     service.AuditActionDelete,
		Metadata:   nil,
	})

	slog.Info("folder deleted", "folder_id", folderID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Folder deleted successfully",
	})
}
