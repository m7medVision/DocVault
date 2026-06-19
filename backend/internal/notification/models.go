package notification

import "time"

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeReminder NotificationType = "reminder"
	NotificationTypeSystem   NotificationType = "system"
)

// NotificationStatus represents the read state of a notification.
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

// Notification represents an in-app notification for a user.
type Notification struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Link      *string                `json:"link,omitempty"`
	Status    NotificationStatus     `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}
