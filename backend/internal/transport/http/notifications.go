package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/notification"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || userID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	status := r.URL.Query().Get("status")
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	input := &usecase.ListNotificationsInput{
		TenantID: tenantID,
		UserID:   userID,
		Status:   notification.NotificationStatus(status),
		Cursor:   cursor,
		Limit:    limit,
	}

	output, err := h.notificationSvc.List(ctx, input)
	if err != nil {
		slog.Error("list notifications failed", "error", err)
		http.Error(w, `{"error":"failed to list notifications","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	unreadCount, _ := h.notificationSvc.GetUnreadCount(ctx, tenantID, userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": output.Notifications,
		"cursor":        output.Cursor,
		"total":         output.Total,
		"unread_count":  unreadCount,
	})
}

func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.HasMinRole(role, middleware.RoleViewer) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" || userID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"notification id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	input := &usecase.MarkReadInput{
		NotificationID: id,
		TenantID:       tenantID,
		UserID:         userID,
	}

	if err := h.notificationSvc.MarkRead(ctx, input); err != nil {
		slog.Error("mark notification read failed", "error", err, "notification_id", id)
		http.Error(w, `{"error":"failed to mark notification as read","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

func (h *Handler) CreateNotificationWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	workerKey := r.Header.Get("X-Worker-API-Key")
	expectedKey := h.cfg.WorkerAPIKey
	if expectedKey == "" {
		http.Error(w, `{"error":"webhook not configured","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}
	if workerKey != expectedKey {
		http.Error(w, `{"error":"invalid API key","code":"UNAUTHORIZED"}`, http.StatusUnauthorized)
		return
	}

	var input struct {
		TenantID string                 `json:"tenant_id"`
		UserID   string                 `json:"user_id"`
		Type     string                 `json:"type"`
		Title    string                 `json:"title"`
		Body     string                 `json:"body"`
		Link     *string                `json:"link,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid request body: %s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if input.TenantID == "" || input.UserID == "" || input.Title == "" {
		http.Error(w, `{"error":"tenant_id, user_id, and title are required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	notifType := input.Type
	if notifType == "" {
		notifType = "reminder"
	}

	notification, err := h.notificationSvc.CreateFromWorker(ctx, &usecase.CreateFromWorkerInput{
		TenantID: input.TenantID,
		UserID:   input.UserID,
		Type:     notifType,
		Title:    input.Title,
		Body:     input.Body,
		Link:     input.Link,
		Metadata: input.Metadata,
	})

	if err != nil {
		slog.Error("create notification webhook failed", "error", err)
		http.Error(w, `{"error":"failed to create notification","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("notification created via webhook",
		"notification_id", notification.ID,
		"user_id", input.UserID,
		"tenant_id", input.TenantID,
	)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      notification.ID,
		"message": "Notification created",
	})
}
