package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	documentapp "github.com/docvault/backend/internal/document/app"
	"github.com/docvault/backend/internal/middleware"
)

// Chat answers questions grounded in a single document's retrieved chunks.
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}
	if h.requireDocVisible(w, r, documentID) {
		return
	}
	h.streamChat(w, r, documentID)
}

// ChatGlobal answers questions grounded in retrieval across all of the caller's
// tenant/org documents (no document scope).
func (h *Handler) ChatGlobal(w http.ResponseWriter, r *http.Request) {
	h.streamChat(w, r, "")
}

// streamChat is the shared implementation for both scoped and global chat.
// documentID is empty for global chat and set for per-document chat.
func (h *Handler) streamChat(w http.ResponseWriter, r *http.Request, documentID string) {
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
		Messages []documentapp.ChatMessage `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if len(body.Messages) == 0 {
		http.Error(w, `{"error":"messages are required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if h.cfg.Search.EmbeddingAPIKey == "" {
		slog.Error("OpenRouter API key not configured")
		http.Error(w, `{"error":"chat service not configured","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	userID := middleware.GetUserID(ctx)
	isAdmin := middleware.HasMinRole(role, middleware.RoleAdmin)
	groupIDs, err := h.aclRepo.ListUserGroupIDs(ctx, userID, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to resolve permissions","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	input := &documentapp.ChatInput{
		DocumentID: documentID,
		Messages:   body.Messages,
		TenantID:   tenantID,
		OrgID:      orgID,
		UserID:     userID,
		GroupIDs:   groupIDs,
		IsAdmin:    isAdmin,
		APIKey:     h.cfg.Search.EmbeddingAPIKey,
		ChatModel:  h.cfg.Search.ChatModel,
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	if err := h.chatSvc.StreamChat(ctx, input, w); err != nil {
		slog.Error("chat stream failed", "error", err, "document_id", documentID, "tenant_id", tenantID)
		errorChunk := map[string]interface{}{
			"type": "RUN_ERROR",
			"error": map[string]string{
				"message": err.Error(),
			},
		}
		chunkBytes, _ := json.Marshal(errorChunk)
		w.Write([]byte("data: " + string(chunkBytes) + "\n\n"))
	}
}
