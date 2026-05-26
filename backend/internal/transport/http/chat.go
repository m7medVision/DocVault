package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
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

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Messages []usecase.ChatMessage `json:"messages"`
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

	input := &usecase.ChatInput{
		DocumentID: documentID,
		Messages:   body.Messages,
		TenantID:   tenantID,
		OrgID:      orgID,
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
