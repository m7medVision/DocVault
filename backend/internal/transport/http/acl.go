package handler

import (
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
)

// requireDocVisible returns true if it already wrote an error response (caller
// should return). Admins short-circuit without a DB lookup. For non-admins it
// evaluates per-document read visibility; an invisible document or any error is
// reported as 404 (not 403) so callers cannot probe for existence.
func (h *Handler) requireDocVisible(w http.ResponseWriter, r *http.Request, documentID string) bool {
	ctx := r.Context()
	role := middleware.GetUserRole(ctx)
	if middleware.HasMinRole(role, middleware.RoleAdmin) {
		return false // short-circuit, no DB
	}
	userID := middleware.GetUserID(ctx)
	groupIDs, _ := h.aclRepo.ListUserGroupIDs(ctx, userID, middleware.GetOrgID(ctx))
	visible, err := h.aclRepo.IsDocumentVisible(ctx, repository.VisibilityParams{
		TenantID:   middleware.GetTenantID(ctx),
		OrgID:      middleware.GetOrgID(ctx),
		DocumentID: documentID,
		UserID:     userID,
		GroupIDs:   groupIDs,
		IsAdmin:    false,
	})
	if err != nil {
		http.Error(w, `{"error":"document not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return true
	}
	if !visible {
		http.Error(w, `{"error":"document not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		return true
	}
	return false
}
