package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	orgID := middleware.GetOrgID(ctx)

	if tenantID == "" || orgID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	members, err := h.membershipRepo.ListByOrg(ctx, tenantID, orgID)
	if err != nil {
		slog.Error("list members failed", "error", err, "tenant_id", tenantID, "org_id", orgID)
		http.Error(w, `{"error":"failed to list members","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
	})
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		http.Error(w, `{"error":"admin permissions required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	var body struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.Email == "" {
		http.Error(w, `{"error":"email is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[body.Role] {
		http.Error(w, `{"error":"invalid role, must be admin, member, or viewer","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	invitationID := generateID("invite")

	h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &userID,
		EntityType: "membership",
		EntityID:   invitationID,
		Action:     usecase.AuditActionCreate,
		Metadata: map[string]interface{}{
			"email": body.Email,
			"role":  body.Role,
		},
	})

	slog.Info("member invitation sent", "invitation_id", invitationID, "email", body.Email, "role", body.Role, "tenant_id", tenantID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      invitationID,
		"email":   body.Email,
		"role":    body.Role,
		"message": "Invitation sent successfully",
	})
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	if tenantID == "" || userID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}
	if h.db == nil || h.authzEnforcer == nil {
		h.writeMissingAuthzDependency(w)
		return
	}

	memberID := r.PathValue("id")
	if memberID == "" {
		http.Error(w, `{"error":"member id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := authz.ValidateRole(body.Role); err != nil || body.Role == authz.RoleOwner {
		http.Error(w, `{"error":"invalid role, must be admin, member, or viewer","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	member, err := h.membershipRepo.GetByID(ctx, tenantID, memberID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"member not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("load member for role update failed", "error", err, "tenant_id", tenantID, "member_id", memberID)
		http.Error(w, `{"error":"failed to load member","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	oldRole := member.Role
	if oldRole == body.Role {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":      memberID,
			"role":    body.Role,
			"message": "Role updated successfully",
		})
		return
	}

	if err := h.membershipRepo.UpdateRole(ctx, memberID, body.Role); err != nil {
		slog.Error("update membership role failed", "error", err, "tenant_id", tenantID, "member_id", memberID)
		http.Error(w, `{"error":"failed to update role","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	if _, err := authz.AssignRoleToUser(h.authzEnforcer, member.UserID, body.Role, tenantID); err != nil {
		if rollbackErr := h.membershipRepo.UpdateRole(ctx, memberID, oldRole); rollbackErr != nil {
			slog.Error("rollback membership role failed", "error", rollbackErr, "tenant_id", tenantID, "member_id", memberID)
		}
		slog.Error("sync casbin role failed", "error", err, "tenant_id", tenantID, "member_id", memberID)
		http.Error(w, `{"error":"failed to update role","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	h.auditSvc.MemberRoleChange(ctx, tenantID, userID, memberID, oldRole, body.Role)

	slog.Info("member role updated", "member_id", memberID, "new_role", body.Role, "tenant_id", tenantID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      memberID,
		"role":    body.Role,
		"message": "Role updated successfully",
	})
}

func (h *Handler) GetMemberPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	actorID := middleware.GetUserID(ctx)

	if tenantID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}
	if h.db == nil || h.authzEnforcer == nil {
		h.writeMissingAuthzDependency(w)
		return
	}

	memberID := r.PathValue("id")
	if memberID == "" {
		http.Error(w, `{"error":"member id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	member, err := h.membershipRepo.GetByID(ctx, tenantID, memberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"member not found","code":"NOT_FOUND"}`, http.StatusNotFound)
			return
		}
		slog.Error("lookup member permissions failed", "error", err, "member_id", memberID, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to load member","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	roles, err := authz.GetRolesForUser(h.authzEnforcer, member.UserID, tenantID)
	if err != nil {
		slog.Error("get member roles failed", "error", err, "member_id", memberID, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to load member permissions","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	permissions, err := authz.GetPermissionsForUser(h.authzEnforcer, member.UserID, tenantID)
	if err != nil {
		slog.Error("get member permissions failed", "error", err, "member_id", memberID, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to load member permissions","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "membership",
		EntityID:   memberID,
		Action:     usecase.AuditActionView,
		Metadata: map[string]interface{}{
			"roles":       roles,
			"permissions": len(permissions),
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"member":       member,
		"roles":        roles,
		"permissions":  permissions,
		"current_role": member.Role,
	})
}

func (h *Handler) writeMissingAuthContext(w http.ResponseWriter) {
	http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
}

func (h *Handler) writeMissingAuthzDependency(w http.ResponseWriter) {
	http.Error(w, `{"error":"authorization backend unavailable","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
