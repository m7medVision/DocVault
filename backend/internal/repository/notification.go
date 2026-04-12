package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotificationNotFound = errors.New("notification not found")

// notificationRepository handles notification data access.
type notificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{db: db}
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

	query := `
		INSERT INTO notifications (id, tenant_id, user_id, type, title, body, link, status, metadata, created_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), $10)
	`
	_, err = r.db.Exec(ctx, query,
		notification.ID, notification.TenantID, notification.UserID,
		notification.Type, notification.Title, notification.Body,
		notification.Link, notification.Status, metadataJSON,
		notification.ReadAt,
	)
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

	query := `
		SELECT id, tenant_id, user_id, type, title, body, link, status, metadata, created_at, read_at
		FROM notifications
		WHERE tenant_id = $1 AND user_id = $2
	`
	args := []interface{}{tenantID, userID}
	argCount := 2

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(status))
	}
	if cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM notifications WHERE id = $%d)", argCount)
		args = append(args, cursor)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]model.Notification, 0)
	for rows.Next() {
		var n model.Notification
		var metadataJSON []byte
		if err := rows.Scan(
			&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Title,
			&n.Body, &n.Link, &n.Status, &metadataJSON,
			&n.CreatedAt, &n.ReadAt,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal notification metadata: %w", err)
			}
		}
		notifications = append(notifications, n)
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
	query := `
		UPDATE notifications
		SET status = 'read', read_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND user_id = $3
	`
	result, err := r.db.Exec(ctx, query, notificationID, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (r *notificationRepository) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND status = 'unread'`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(&count)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}
