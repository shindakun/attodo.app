package models

import "time"

// NotificationUser represents a user who has enabled notifications
type NotificationUser struct {
	DID                  string     `db:"did" json:"did"`
	NotificationsEnabled bool       `db:"notifications_enabled" json:"notificationsEnabled"`
	LastCheckedAt        *time.Time `db:"last_checked_at" json:"lastCheckedAt,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
}

// PushSubscription represents a Web Push subscription for a user's device/browser
type PushSubscription struct {
	ID         int64     `db:"id" json:"id"`
	DID        string    `db:"did" json:"did"`
	Endpoint   string    `db:"endpoint" json:"endpoint"`
	P256dhKey  string    `db:"p256dh_key" json:"p256dhKey"`
	AuthSecret string    `db:"auth_secret" json:"authSecret"`
	UserAgent  string    `db:"user_agent" json:"userAgent,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	LastUsedAt time.Time `db:"last_used_at" json:"lastUsedAt"`
}

// NotificationHistory tracks sent notifications to prevent spam
type NotificationHistory struct {
	ID               int64     `db:"id" json:"id"`
	DID              string    `db:"did" json:"did"`
	TaskURI          string    `db:"task_uri" json:"taskUri"`
	NotificationType string    `db:"notification_type" json:"notificationType"` // 'overdue', 'due_today', 'due_soon'
	SentAt           time.Time `db:"sent_at" json:"sentAt"`
	Status           string    `db:"status" json:"status"` // 'sent', 'failed', 'expired'
	ErrorMessage     string    `db:"error_message" json:"errorMessage,omitempty"`
}
