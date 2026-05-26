package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		http.Error(w, `{"error":"admin permissions required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	actorID := r.URL.Query().Get("actor_id")
	action := r.URL.Query().Get("action")
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	input := &usecase.ListAuditEventsInput{
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   entityID,
		ActorID:    actorID,
		Action:     action,
		Cursor:     cursor,
		Limit:      limit,
	}

	output, err := h.auditSvc.List(ctx, input)
	if err != nil {
		slog.Error("list audit events failed", "error", err, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to list audit events","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": output.Events,
		"cursor": output.Cursor,
	})
}
