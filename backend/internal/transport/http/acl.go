package handler

import (
	"encoding/json"
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
)

var validACLPermissions = map[string]bool{"read": true, "write": true, "delete": true}
var validPrincipalTypes = map[string]bool{"user": true, "group": true}
var validResourceTypes = map[string]bool{"document": true, "folder": true}

// principalFrom builds the caller's security context from the request. IsAdmin
// captures the admin short-circuit (admin or owner) so the app-layer Authorizer
// stays free of the role hierarchy.
func principalFrom(r *http.Request) usecase.Principal {
	ctx := r.Context()
	role := middleware.GetUserRole(ctx)
	return usecase.Principal{
		TenantID: middleware.GetTenantID(ctx),
		OrgID:    middleware.GetOrgID(ctx),
		UserID:   middleware.GetUserID(ctx),
		Role:     role,
		IsAdmin:  middleware.HasMinRole(role, middleware.RoleAdmin),
	}
}

// requireDocVisible returns true if it already wrote an error response (caller
// should return). It delegates the decision to the app-layer Authorizer, which
// owns the admin short-circuit and the invisible/error -> 404 (not 403) rule.
func (h *Handler) requireDocVisible(w http.ResponseWriter, r *http.Request, documentID string) bool {
	if err := h.authz.RequireDocVisible(r.Context(), principalFrom(r), documentID); err != nil {
		writeErr(w, r, err)
		return true
	}
	return false
}

// requireFolderVisible returns true if it already wrote an error response
// (caller should return). It delegates to the Authorizer, which owns the admin
// short-circuit and the invisible/error -> 404 (not 403) rule.
func (h *Handler) requireFolderVisible(w http.ResponseWriter, r *http.Request, folderID string) bool {
	if err := h.authz.RequireFolderVisible(r.Context(), principalFrom(r), folderID); err != nil {
		writeErr(w, r, err)
		return true
	}
	return false
}

// requireDocWritable returns true if it already wrote an error response (caller
// should return). It delegates to the Authorizer, which owns the admin
// short-circuit and the non-writable/error -> 404 (not 403) rule so a read-only
// grant cannot be used to probe for or mutate a restricted document.
func (h *Handler) requireDocWritable(w http.ResponseWriter, r *http.Request, documentID string) bool {
	if err := h.authz.RequireDocWritable(r.Context(), principalFrom(r), documentID); err != nil {
		writeErr(w, r, err)
		return true
	}
	return false
}

// setDocumentRestricted toggles a document's restriction flag.
func (h *Handler) setDocumentRestricted(w http.ResponseWriter, r *http.Request, restricted bool) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	rows, err := h.aclRepo.SetDocumentRestricted(ctx, repository.SetRestrictedParams{
		TenantID:   tenantID,
		OrgID:      orgID,
		ResourceID: documentID,
		Restricted: restricted,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to update document restriction","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":"document not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "document",
		EntityID:   documentID,
		Action:     usecase.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":     "restrict",
			"restricted": restricted,
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         documentID,
		"restricted": restricted,
	})
}

// SetDocumentRestricted marks a document as restricted.
func (h *Handler) SetDocumentRestricted(w http.ResponseWriter, r *http.Request) {
	h.setDocumentRestricted(w, r, true)
}

// UnsetDocumentRestricted clears a document's restriction.
func (h *Handler) UnsetDocumentRestricted(w http.ResponseWriter, r *http.Request) {
	h.setDocumentRestricted(w, r, false)
}

// setFolderRestricted toggles a folder's restriction flag.
func (h *Handler) setFolderRestricted(w http.ResponseWriter, r *http.Request, restricted bool) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	folderID := r.PathValue("id")
	if folderID == "" {
		http.Error(w, `{"error":"folder id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	rows, err := h.aclRepo.SetFolderRestricted(ctx, repository.SetRestrictedParams{
		TenantID:   tenantID,
		OrgID:      orgID,
		ResourceID: folderID,
		Restricted: restricted,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to update folder restriction","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":"folder not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "folder",
		EntityID:   folderID,
		Action:     usecase.AuditActionUpdate,
		Metadata: map[string]interface{}{
			"action":     "restrict",
			"restricted": restricted,
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         folderID,
		"restricted": restricted,
	})
}

// SetFolderRestricted marks a folder as restricted.
func (h *Handler) SetFolderRestricted(w http.ResponseWriter, r *http.Request) {
	h.setFolderRestricted(w, r, true)
}

// UnsetFolderRestricted clears a folder's restriction.
func (h *Handler) UnsetFolderRestricted(w http.ResponseWriter, r *http.Request) {
	h.setFolderRestricted(w, r, false)
}

type createGrantRequest struct {
	ResourceType  string `json:"resource_type"`
	ResourceID    string `json:"resource_id"`
	PrincipalType string `json:"principal_type"`
	PrincipalID   string `json:"principal_id"`
	Permission    string `json:"permission"`
}

// CreateGrant attaches a read/write/delete grant to a document or folder for a
// user or group principal.
func (h *Handler) CreateGrant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	var body createGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.ResourceID == "" || body.PrincipalID == "" {
		http.Error(w, `{"error":"resource_id and principal_id are required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if !validResourceTypes[body.ResourceType] {
		http.Error(w, `{"error":"invalid resource_type","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if !validPrincipalTypes[body.PrincipalType] {
		http.Error(w, `{"error":"invalid principal_type","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if body.Permission == "" {
		body.Permission = "read"
	}
	if !validACLPermissions[body.Permission] {
		http.Error(w, `{"error":"invalid permission","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	exists, err := h.aclRepo.GrantTargetExists(ctx, repository.ResourceRef{
		TenantID:     tenantID,
		OrgID:        orgID,
		ResourceType: body.ResourceType,
		ResourceID:   body.ResourceID,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to create grant","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, `{"error":"resource not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	id, err := h.aclRepo.CreateGrant(ctx, repository.CreateGrantParams{
		TenantID:      tenantID,
		OrgID:         orgID,
		ResourceType:  body.ResourceType,
		ResourceID:    body.ResourceID,
		PrincipalType: body.PrincipalType,
		PrincipalID:   body.PrincipalID,
		Permission:    body.Permission,
		GrantedBy:     &actorID,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to create grant","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "acl_grant",
		EntityID:   body.ResourceID,
		Action:     usecase.AuditActionShare,
		Metadata: map[string]interface{}{
			"resource_type":  body.ResourceType,
			"resource_id":    body.ResourceID,
			"principal_type": body.PrincipalType,
			"principal_id":   body.PrincipalID,
			"permission":     body.Permission,
		},
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             id,
		"resource_type":  body.ResourceType,
		"resource_id":    body.ResourceID,
		"principal_type": body.PrincipalType,
		"principal_id":   body.PrincipalID,
		"permission":     body.Permission,
	})
}

// DeleteGrant removes a grant by id.
func (h *Handler) DeleteGrant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	grantID := r.PathValue("id")
	if grantID == "" {
		http.Error(w, `{"error":"grant id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	rows, err := h.aclRepo.DeleteGrant(ctx, tenantID, orgID, grantID)
	if err != nil {
		http.Error(w, `{"error":"failed to delete grant","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":"grant not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "acl_grant",
		EntityID:   grantID,
		Action:     usecase.AuditActionDelete,
		Metadata:   nil,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "grant removed successfully"})
}

// ListGrants lists the grants attached to a single resource.
func (h *Handler) ListGrants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)
	actorID := middleware.GetUserID(ctx)
	if tenantID == "" || orgID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}

	resourceType := r.URL.Query().Get("resource_type")
	resourceID := r.URL.Query().Get("resource_id")
	if !validResourceTypes[resourceType] {
		http.Error(w, `{"error":"invalid resource_type","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if resourceID == "" {
		http.Error(w, `{"error":"resource_id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	grants, err := h.aclRepo.ListGrantsByResource(ctx, repository.ResourceRef{
		TenantID:     tenantID,
		OrgID:        orgID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to list grants","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"grants": grants,
	})
}
