package handler

import (
	"encoding/json"
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
)

type createGroupRequest struct {
	Name string `json:"name"`
}

// CreateGroup creates an org-scoped principal group.
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	var body createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, `{"error":"name is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	group, err := h.aclRepo.CreateGroup(ctx, repository.CreateGroupParams{
		TenantID:  tenantID,
		OrgID:     orgID,
		Name:      body.Name,
		CreatedBy: &actorID,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to create group","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "group",
		EntityID:   group.ID,
		Action:     usecase.AuditActionCreate,
		Metadata: map[string]interface{}{
			"name": group.Name,
		},
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"group": group,
	})
}

// DeleteGroup deletes an org-scoped group.
func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	groupID := r.PathValue("id")
	if groupID == "" {
		http.Error(w, `{"error":"group id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	rows, err := h.aclRepo.DeleteGroup(ctx, tenantID, orgID, groupID)
	if err != nil {
		http.Error(w, `{"error":"failed to delete group","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":"group not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "group",
		EntityID:   groupID,
		Action:     usecase.AuditActionDelete,
		Metadata:   nil,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "group removed successfully"})
}

// ListGroups returns the groups defined within the org.
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	groups, err := h.aclRepo.ListGroups(ctx, tenantID, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to list groups","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
	})
}

type addGroupMemberRequest struct {
	UserID string `json:"user_id"`
}

// AddGroupMember adds a user to a group.
func (h *Handler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	groupID := r.PathValue("id")
	if groupID == "" {
		http.Error(w, `{"error":"group id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var body addGroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if body.UserID == "" {
		http.Error(w, `{"error":"user_id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := h.aclRepo.AddGroupMember(ctx, tenantID, orgID, groupID, body.UserID); err != nil {
		http.Error(w, `{"error":"failed to add group member","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "group",
		EntityID:   groupID,
		Action:     usecase.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":  "add_member",
			"user_id": body.UserID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"group_id": groupID,
		"user_id":  body.UserID,
		"message":  "member added successfully",
	})
}

// RemoveGroupMember removes a user from a group.
func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	groupID := r.PathValue("id")
	userID := r.PathValue("user_id")
	if groupID == "" || userID == "" {
		http.Error(w, `{"error":"group id and user id are required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	rows, err := h.aclRepo.RemoveGroupMember(ctx, tenantID, orgID, groupID, userID)
	if err != nil {
		http.Error(w, `{"error":"failed to remove group member","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":"group member not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "group",
		EntityID:   groupID,
		Action:     usecase.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":  "remove_member",
			"user_id": userID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed successfully"})
}
