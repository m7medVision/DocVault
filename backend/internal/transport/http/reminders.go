package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) ListReminders(w http.ResponseWriter, r *http.Request) {
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

	documentID := r.URL.Query().Get("document_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	input := &usecase.ListRemindersInput{
		TenantID:   tenantID,
		DocumentID: documentID,
		ActiveOnly: activeOnly,
	}

	output, err := h.reminderSvc.List(ctx, input)
	if err != nil {
		slog.Error("list reminders failed", "error", err)
		http.Error(w, `{"error":"failed to list reminders","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"reminders": output.Reminders,
	})
}

func (h *Handler) CreateReminder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, `{"error":"document id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var input struct {
		RuleType         string `json:"rule_type"`
		TriggerDate      string `json:"trigger_date"`
		NotifyDaysBefore []int  `json:"notify_days_before"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid request body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	triggerDate, err := time.Parse(time.RFC3339, input.TriggerDate)
	if err != nil {
		http.Error(w, `{"error":"invalid trigger_date format, use RFC3339","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	createInput := &usecase.CreateReminderInput{
		TenantID:         tenantID,
		DocumentID:       documentID,
		RuleType:         input.RuleType,
		TriggerDate:      triggerDate,
		NotifyDaysBefore: input.NotifyDaysBefore,
		Source:           "manual",
	}

	output, err := h.reminderSvc.Create(ctx, createInput)
	if err != nil {
		slog.Error("create reminder failed", "error", err)
		http.Error(w, `{"error":"failed to create reminder","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"reminder": output.Reminder,
	})
}

func (h *Handler) UpdateReminder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	role := middleware.GetUserRole(ctx)

	if !middleware.CanWrite(role) {
		http.Error(w, `{"error":"insufficient permissions","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	if tenantID == "" {
		http.Error(w, `{"error":"tenant context required","code":"FORBIDDEN"}`, http.StatusForbidden)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"reminder id is required","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	var input struct {
		Active *bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid request body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	updateInput := &usecase.UpdateReminderInput{
		TenantID: tenantID,
		ID:       id,
		Active:   input.Active,
	}

	output, err := h.reminderSvc.Update(ctx, updateInput)
	if err != nil {
		slog.Error("update reminder failed", "error", err)
		http.Error(w, `{"error":"failed to update reminder","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"reminder": output.Reminder,
	})
}
