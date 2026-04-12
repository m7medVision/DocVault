package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/service"
)

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	input := &service.ListTagsInput{
		TenantID: tenantID,
		Query:    query,
		Limit:    limit,
	}

	output, err := h.tagSvc.List(ctx, input)
	if err != nil {
		slog.Error("list tags failed", "error", err, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to list tags","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tags": output.Tags,
	})
}
