package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/notification"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotificationNotFound = errors.New("notification not found")

// notificationRepository handles notification data access.
type notificationRepository struct {
	queries sqldb.Querier
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{queries: sqldb.New(db)}
}

// Create creates a new notification.
func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification is nil")
	}

	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal notification metadata: %w", err)
	}

	err = r.queries.CreateNotification(ctx, sqldb.CreateNotificationParams{
		ID:       notification.ID,
		TenantID: notification.TenantID,
		UserID:   notification.UserID,
		Type:     string(notification.Type),
		Title:    notification.Title,
		Body:     &notification.Body,
		Link:     notification.Link,
		Status:   string(notification.Status),
		Metadata: metadataJSON,
		ReadAt:   timestamptzFromTimePtr(notification.ReadAt),
	})
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// List returns notifications for a user with cursor pagination.
func (r *notificationRepository) List(ctx context.Context, tenantID, userID string, status model.NotificationStatus, cursor string, limit int) ([]model.Notification, *string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var statusFilter *string
	if status != "" {
		value := string(status)
		statusFilter = &value
	}
	var cursorFilter *string
	if cursor != "" {
		cursorFilter = &cursor
	}

	notificationRows, err := r.queries.ListNotifications(ctx, sqldb.ListNotificationsParams{
		TenantIDArg: tenantID,
		UserIDArg:   userID,
		Status:      statusFilter,
		CursorID:    cursorFilter,
		LimitCount:  int32(limit + 1),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	notifications := make([]model.Notification, 0, len(notificationRows))
	for _, row := range notificationRows {
		notification, err := toModelNotification(row)
		if err != nil {
			return nil, nil, err
		}
		notifications = append(notifications, notification)
	}

	var resultCursor *string
	if len(notifications) > limit {
		notifications = notifications[:limit]
		c := notifications[len(notifications)-1].ID
		resultCursor = &c
	}

	return notifications, resultCursor, nil
}

// MarkRead marks a notification as read.
func (r *notificationRepository) MarkRead(ctx context.Context, tenantID, userID, notificationID string) error {
	rowsAffected, err := r.queries.MarkNotificationRead(ctx, sqldb.MarkNotificationReadParams{
		ID:       notificationID,
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (r *notificationRepository) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	count, err := r.queries.GetUnreadNotificationCount(ctx, sqldb.GetUnreadNotificationCountParams{TenantID: tenantID, UserID: userID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return int(count), nil
}

func toModelNotification(notification sqldb.Notification) (model.Notification, error) {
	modelNotification := model.Notification{
		ID:        notification.ID,
		TenantID:  notification.TenantID,
		UserID:    notification.UserID,
		Type:      model.NotificationType(notification.Type),
		Title:     notification.Title,
		Link:      notification.Link,
		Status:    model.NotificationStatus(notification.Status),
		CreatedAt: notification.CreatedAt.Time,
		ReadAt:    timePtrFromTimestamptz(notification.ReadAt),
	}
	if notification.Body != nil {
		modelNotification.Body = *notification.Body
	}
	if len(notification.Metadata) > 0 {
		if err := json.Unmarshal(notification.Metadata, &modelNotification.Metadata); err != nil {
			return model.Notification{}, fmt.Errorf("failed to unmarshal notification metadata: %w", err)
		}
	}
	return modelNotification, nil
}
