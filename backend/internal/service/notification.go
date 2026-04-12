// Package service provides business logic for DocVault.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

// NotificationService handles notification operations.
type NotificationService struct {
	repo repository.NotificationRepository
}

// NewNotificationService creates a new notification service.
func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// ListNotificationsInput is the input for listing notifications.
type ListNotificationsInput struct {
	TenantID string
	UserID   string
	Status   model.NotificationStatus
	Limit    int
	Cursor   string
}

// ListNotificationsOutput is the output from listing notifications.
type ListNotificationsOutput struct {
	Notifications []*model.Notification
	Cursor        string
	Total         int
}

var errTenantAndUserIDRequired = errors.New("tenant_id and user_id are required")

// List returns notifications for a user.
func (s *NotificationService) List(ctx context.Context, input *ListNotificationsInput) (*ListNotificationsOutput, error) {
	if input.TenantID == "" || input.UserID == "" {
		return nil, errTenantAndUserIDRequired
	}

	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if s.repo == nil {
		return &ListNotificationsOutput{
			Notifications: []*model.Notification{},
			Total:         0,
		}, nil
	}

	notifications, cursor, err := s.repo.List(ctx, input.TenantID, input.UserID, input.Status, input.Cursor, limit)
	if err != nil {
		return nil, err
	}

	ptrNotifications := make([]*model.Notification, len(notifications))
	for i := range notifications {
		ptrNotifications[i] = &notifications[i]
	}

	resultCursor := ""
	if cursor != nil {
		resultCursor = *cursor
	}

	slog.Info("notifications listed",
		"user_id", input.UserID,
		"tenant_id", input.TenantID,
		"count", len(ptrNotifications),
	)

	return &ListNotificationsOutput{
		Notifications: ptrNotifications,
		Cursor:        resultCursor,
		Total:         len(ptrNotifications),
	}, nil
}

// MarkReadInput is the input for marking a notification as read.
type MarkReadInput struct {
	NotificationID string
	TenantID       string
	UserID         string
}

var errNotificationIDRequired = errors.New("notification_id is required")

// MarkRead marks a notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, input *MarkReadInput) error {
	if input.NotificationID == "" {
		return errNotificationIDRequired
	}

	if s.repo == nil {
		return errors.New("notification service not fully implemented")
	}

	if err := s.repo.MarkRead(ctx, input.TenantID, input.UserID, input.NotificationID); err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}

	slog.Info("notification marked as read",
		"notification_id", input.NotificationID,
		"user_id", input.UserID,
	)

	return nil
}

// CreateReminderNotification creates a reminder notification for a user.
func (s *NotificationService) CreateReminderNotification(
	ctx context.Context,
	tenantID, userID, documentID string,
	ruleType string,
	triggerDate time.Time,
	daysUntil int,
) (*model.Notification, error) {
	var title, body string
	switch ruleType {
	case "expiry":
		title = "Document Expiring Soon"
		body = fmt.Sprintf("A document will expire in %d days", daysUntil)
	case "renewal":
		title = "Document Renewal Required"
		body = fmt.Sprintf("A document requires renewal in %d days", daysUntil)
	case "due_date":
		title = "Document Due Date Approaching"
		body = fmt.Sprintf("A document is due in %d days", daysUntil)
	default:
		title = "Document Reminder"
		body = fmt.Sprintf("You have a reminder for an upcoming document (%d days)", daysUntil)
	}

	link := fmt.Sprintf("/documents/%s", documentID)

	notification := &model.Notification{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		UserID:   userID,
		Type:     model.NotificationTypeReminder,
		Title:    title,
		Body:     body,
		Link:     &link,
		Status:   model.NotificationStatusUnread,
		Metadata: map[string]interface{}{
			"document_id":  documentID,
			"rule_type":    ruleType,
			"trigger_date": triggerDate.Format(time.RFC3339),
			"days_until":   daysUntil,
		},
		CreatedAt: time.Now(),
	}

	if s.repo != nil {
		if err := s.repo.Create(ctx, notification); err != nil {
			return nil, err
		}
	}

	slog.Info("reminder notification created",
		"notification_id", notification.ID,
		"user_id", userID,
		"tenant_id", tenantID,
		"document_id", documentID,
		"rule_type", ruleType,
	)

	return notification, nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (s *NotificationService) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	if s.repo == nil {
		return 0, nil
	}

	count, err := s.repo.GetUnreadCount(ctx, tenantID, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

var errWorkerInputRequired = errors.New("tenant_id, user_id, and title are required")

// CreateFromWorkerInput is the input for creating a notification from the worker.
type CreateFromWorkerInput struct {
	TenantID string
	UserID   string
	Type     string
	Title    string
	Body     string
	Link     *string
	Metadata map[string]interface{}
}

// CreateFromWorker creates a notification from worker webhook.
func (s *NotificationService) CreateFromWorker(ctx context.Context, input *CreateFromWorkerInput) (*model.Notification, error) {
	if input.TenantID == "" || input.UserID == "" || input.Title == "" {
		return nil, errWorkerInputRequired
	}

	notification := &model.Notification{
		ID:        uuid.New().String(),
		TenantID:  input.TenantID,
		UserID:    input.UserID,
		Type:      model.NotificationType(input.Type),
		Title:     input.Title,
		Body:      input.Body,
		Link:      input.Link,
		Status:    model.NotificationStatusUnread,
		Metadata:  input.Metadata,
		CreatedAt: time.Now(),
	}

	if s.repo != nil {
		if err := s.repo.Create(ctx, notification); err != nil {
			return nil, err
		}
	}

	slog.Info("notification created from worker",
		"notification_id", notification.ID,
		"user_id", input.UserID,
		"tenant_id", input.TenantID,
	)

	return notification, nil
}
