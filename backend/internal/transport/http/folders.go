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

	input := &documentapp.CreateFolderInput{
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

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   output.Folder.ID,
		Action:     auditapp.AuditActionCreate,
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

	input := &documentapp.ListFoldersInput{
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

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     auditapp.AuditActionUpdate,
		Metadata:   map[string]interface{}{"action": "rename", "name": body.Name},
	})

	slog.Info("folder renamed", "folder_id", folderID, "name", body.Name, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":   folderID,
		"name": body.Name,
	})
}

// GetFolderIndex returns a folder's optional markdown "About this folder"
// overview. Beyond the route-level Casbin folders/read check it enforces
// per-folder read visibility via requireFolderVisible: a restricted folder a
// caller cannot see returns 404 (never leaking the overview). Response:
// {"index_content": <string|null>}.
func (h *Handler) GetFolderIndex(w http.ResponseWriter, r *http.Request) {
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

	folderID := r.PathValue("id")
	if folderID == "" {
		http.Error(w, `{"error":"folder id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.requireFolderVisible(w, r, folderID) {
		return
	}

	content, err := h.folderSvc.GetIndex(ctx, tenantID, orgID, folderID)
	if err != nil {
		if errors.Is(err, documentapp.ErrFolderNotFound) {
			http.Error(w, `{"error":"folder not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("get folder index failed", "error", err, "folder_id", folderID)
		http.Error(w, `{"error":"failed to get folder index","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"index_content": content,
	})
}

// SetFolderIndex updates a folder's markdown "About this folder" overview.
// Beyond the route-level Casbin folders/write check it enforces per-folder read
// visibility via requireFolderVisible: a restricted folder a caller cannot see
// returns 404. Body: {"index_content": string}.
func (h *Handler) SetFolderIndex(w http.ResponseWriter, r *http.Request) {
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

	if h.requireFolderVisible(w, r, folderID) {
		return
	}

	var body struct {
		IndexContent string `json:"index_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	content := body.IndexContent
	if err := h.folderSvc.SetIndex(ctx, tenantID, orgID, folderID, &content); err != nil {
		if errors.Is(err, documentapp.ErrFolderNotFound) {
			http.Error(w, `{"error":"folder not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("set folder index failed", "error", err, "folder_id", folderID)
		http.Error(w, `{"error":"failed to set folder index","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     auditapp.AuditActionUpdate,
		Metadata:   map[string]interface{}{"action": "set_index"},
	})

	slog.Info("folder index updated", "folder_id", folderID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": folderID,
	})
}

// MoveFolder reparents a folder, preserving its name. Body: {"parent_id":
// "<uuid>"|null}. A null parent_id moves the folder to root. Self-parent,
// cycle (moving under itself or a descendant), and depth-cap violations are
// rejected with 400 BAD_REQUEST; a missing target parent yields 404.
func (h *Handler) MoveFolder(w http.ResponseWriter, r *http.Request) {
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
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := h.folderSvc.Reparent(ctx, tenantID, orgID, folderID, body.ParentID); err != nil {
		switch {
		case errors.Is(err, documentapp.ErrFolderNotFound):
			http.Error(w, `{"error":"folder not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		case errors.Is(err, documentapp.ErrTargetParentNotFound):
			http.Error(w, `{"error":"target parent folder not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		case errors.Is(err, documentapp.ErrFolderSelfParent):
			http.Error(w, `{"error":"a folder cannot be its own parent","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		case errors.Is(err, documentapp.ErrFolderCycle):
			http.Error(w, `{"error":"cannot move a folder into itself or one of its descendants","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		case errors.Is(err, documentapp.ErrFolderDepthExceeded):
			http.Error(w, `{"error":"resulting folder depth exceeds the maximum allowed","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		default:
			slog.Error("move folder failed", "error", err, "folder_id", folderID)
			http.Error(w, `{"error":"failed to move folder","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		}
		return
	}

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     auditapp.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":    "move",
			"parent_id": body.ParentID,
		},
	})

	slog.Info("folder moved", "folder_id", folderID, "parent_id", body.ParentID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":        folderID,
		"parent_id": body.ParentID,
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

	h.auditSvc.Write(ctx, &auditapp.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     auditapp.AuditActionDelete,
		Metadata:   nil,
	})

	slog.Info("folder deleted", "folder_id", folderID, "tenant_id", tenantID, "actor_id", userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Folder deleted successfully",
	})
}
