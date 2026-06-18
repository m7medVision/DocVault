package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Query     string   `json:"query"`
		Limit     int      `json:"limit"`
		DocType   string   `json:"doc_type"`
		Language  string   `json:"language"`
		Status    string   `json:"status"`
		FolderID  string   `json:"folder_id"`
		Tags      []string `json:"tags"`
		StartDate string   `json:"start_date"`
		EndDate   string   `json:"end_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if body.Query == "" {
		http.Error(w, `{"error":"query is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(ctx)
	isAdmin := middleware.HasMinRole(role, middleware.RoleAdmin)
	groupIDs, _ := h.aclRepo.ListUserGroupIDs(ctx, userID, orgID)

	input := &usecase.SearchInput{
		Query:     body.Query,
		Limit:     body.Limit,
		DocType:   body.DocType,
		Language:  body.Language,
		Status:    body.Status,
		FolderID:  body.FolderID,
		Tags:      body.Tags,
		StartDate: body.StartDate,
		EndDate:   body.EndDate,
		TenantID:  tenantID,
		OrgID:     orgID,
		UserID:    userID,
		GroupIDs:  groupIDs,
		IsAdmin:   isAdmin,
	}

	output, err := h.searchSvc.Search(ctx, input)
	if err != nil {
		slog.Error("search failed", "error", err, "query", body.Query)
		http.Error(w, fmt.Sprintf(`{"error":"search failed: %s","code":"INTERNAL_ERROR"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	slog.Info("search request", "query", body.Query, "tenant_id", tenantID, "org_id", orgID, "results", output.Total)

	writeJSON(w, http.StatusOK, output)
}
