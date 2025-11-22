package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shindakun/attodo/internal/models"
)

// NotificationRepo handles database operations for notifications
type NotificationRepo struct {
	db *DB
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo(db *DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// ============================================================================
// NOTIFICATION USERS
// ============================================================================

// GetNotificationUser retrieves a notification user by DID
func (r *NotificationRepo) GetNotificationUser(did string) (*models.NotificationUser, error) {
	var user models.NotificationUser
	err := r.db.QueryRow(`
		SELECT did, notifications_enabled, last_checked_at, created_at, updated_at
		FROM notification_users
		WHERE did = ?
	`, did).Scan(
		&user.DID,
		&user.NotificationsEnabled,
		&user.LastCheckedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification user: %w", err)
	}

	return &user, nil
}

// CreateNotificationUser creates a new notification user
func (r *NotificationRepo) CreateNotificationUser(user *models.NotificationUser) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(`
		INSERT INTO notification_users (did, notifications_enabled, last_checked_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.DID, user.NotificationsEnabled, user.LastCheckedAt, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification user: %w", err)
	}

	return nil
}

// UpdateNotificationUser updates an existing notification user
func (r *NotificationRepo) UpdateNotificationUser(user *models.NotificationUser) error {
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		UPDATE notification_users
		SET notifications_enabled = ?, last_checked_at = ?, updated_at = ?
		WHERE did = ?
	`, user.NotificationsEnabled, user.LastCheckedAt, user.UpdatedAt, user.DID)

	if err != nil {
		return fmt.Errorf("failed to update notification user: %w", err)
	}

	return nil
}

// GetEnabledNotificationUsers retrieves all users with notifications enabled
func (r *NotificationRepo) GetEnabledNotificationUsers() ([]*models.NotificationUser, error) {
	rows, err := r.db.Query(`
		SELECT did, notifications_enabled, last_checked_at, created_at, updated_at
		FROM notification_users
		WHERE notifications_enabled = 1
		ORDER BY last_checked_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled users: %w", err)
	}
	defer rows.Close()

	var users []*models.NotificationUser
	for rows.Next() {
		var user models.NotificationUser
		if err := rows.Scan(
			&user.DID,
			&user.NotificationsEnabled,
			&user.LastCheckedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading notification users: %w", err)
	}

	return users, nil
}

// ============================================================================
// PUSH SUBSCRIPTIONS
// ============================================================================

// GetPushSubscription retrieves a push subscription by endpoint
func (r *NotificationRepo) GetPushSubscription(endpoint string) (*models.PushSubscription, error) {
	var sub models.PushSubscription
	err := r.db.QueryRow(`
		SELECT id, did, endpoint, p256dh_key, auth_secret, user_agent, created_at, last_used_at
		FROM push_subscriptions
		WHERE endpoint = ?
	`, endpoint).Scan(
		&sub.ID,
		&sub.DID,
		&sub.Endpoint,
		&sub.P256dhKey,
		&sub.AuthSecret,
		&sub.UserAgent,
		&sub.CreatedAt,
		&sub.LastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get push subscription: %w", err)
	}

	return &sub, nil
}

// GetPushSubscriptionsByDID retrieves all push subscriptions for a DID
func (r *NotificationRepo) GetPushSubscriptionsByDID(did string) ([]*models.PushSubscription, error) {
	rows, err := r.db.Query(`
		SELECT id, did, endpoint, p256dh_key, auth_secret, user_agent, created_at, last_used_at
		FROM push_subscriptions
		WHERE did = ?
		ORDER BY created_at DESC
	`, did)
	if err != nil {
		return nil, fmt.Errorf("failed to query push subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		if err := rows.Scan(
			&sub.ID,
			&sub.DID,
			&sub.Endpoint,
			&sub.P256dhKey,
			&sub.AuthSecret,
			&sub.UserAgent,
			&sub.CreatedAt,
			&sub.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan push subscription: %w", err)
		}
		subs = append(subs, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading push subscriptions: %w", err)
	}

	return subs, nil
}

// CreatePushSubscription creates a new push subscription
func (r *NotificationRepo) CreatePushSubscription(sub *models.PushSubscription) error {
	now := time.Now()
	sub.CreatedAt = now
	sub.LastUsedAt = now

	result, err := r.db.Exec(`
		INSERT INTO push_subscriptions (did, endpoint, p256dh_key, auth_secret, user_agent, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sub.DID, sub.Endpoint, sub.P256dhKey, sub.AuthSecret, sub.UserAgent, sub.CreatedAt, sub.LastUsedAt)

	if err != nil {
		return fmt.Errorf("failed to create push subscription: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get subscription ID: %w", err)
	}

	sub.ID = id
	return nil
}

// UpdatePushSubscriptionLastUsed updates the last_used_at timestamp
func (r *NotificationRepo) UpdatePushSubscriptionLastUsed(endpoint string) error {
	_, err := r.db.Exec(`
		UPDATE push_subscriptions
		SET last_used_at = ?
		WHERE endpoint = ?
	`, time.Now(), endpoint)

	if err != nil {
		return fmt.Errorf("failed to update subscription last used: %w", err)
	}

	return nil
}

// DeletePushSubscription deletes a push subscription by endpoint
func (r *NotificationRepo) DeletePushSubscription(endpoint string) error {
	_, err := r.db.Exec(`
		DELETE FROM push_subscriptions
		WHERE endpoint = ?
	`, endpoint)

	if err != nil {
		return fmt.Errorf("failed to delete push subscription: %w", err)
	}

	return nil
}

// DeletePushSubscriptionsByDID deletes all push subscriptions for a DID
func (r *NotificationRepo) DeletePushSubscriptionsByDID(did string) error {
	_, err := r.db.Exec(`
		DELETE FROM push_subscriptions
		WHERE did = ?
	`, did)

	if err != nil {
		return fmt.Errorf("failed to delete push subscriptions: %w", err)
	}

	return nil
}

// ============================================================================
// NOTIFICATION HISTORY
// ============================================================================

// CreateNotificationHistory records a sent notification
func (r *NotificationRepo) CreateNotificationHistory(history *models.NotificationHistory) error {
	history.SentAt = time.Now()

	result, err := r.db.Exec(`
		INSERT INTO notification_history (did, task_uri, notification_type, sent_at, status, error_message)
		VALUES (?, ?, ?, ?, ?, ?)
	`, history.DID, history.TaskURI, history.NotificationType, history.SentAt, history.Status, history.ErrorMessage)

	if err != nil {
		return fmt.Errorf("failed to create notification history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get history ID: %w", err)
	}

	history.ID = id
	return nil
}

// GetRecentNotification checks if we recently sent a notification for this task
// Returns the most recent notification within the cooldown period
func (r *NotificationRepo) GetRecentNotification(did, taskURI string, cooldownHours int) (*models.NotificationHistory, error) {
	cutoff := time.Now().Add(-time.Duration(cooldownHours) * time.Hour)

	var history models.NotificationHistory
	err := r.db.QueryRow(`
		SELECT id, did, task_uri, notification_type, sent_at, status, error_message
		FROM notification_history
		WHERE did = ? AND task_uri = ? AND sent_at > ? AND status = 'sent'
		ORDER BY sent_at DESC
		LIMIT 1
	`, did, taskURI, cutoff).Scan(
		&history.ID,
		&history.DID,
		&history.TaskURI,
		&history.NotificationType,
		&history.SentAt,
		&history.Status,
		&history.ErrorMessage,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No recent notification
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recent notification: %w", err)
	}

	return &history, nil
}

// CleanupOldHistory deletes notification history older than the specified days
func (r *NotificationRepo) CleanupOldHistory(daysToKeep int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -daysToKeep)

	result, err := r.db.Exec(`
		DELETE FROM notification_history
		WHERE sent_at < ?
	`, cutoff)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old history: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}
