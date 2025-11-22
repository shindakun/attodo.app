-- Initial database schema for notification system
-- Phase 2: Server-side tracking and push subscriptions

-- Users who have enabled notifications
CREATE TABLE IF NOT EXISTS notification_users (
    did TEXT PRIMARY KEY,
    notifications_enabled BOOLEAN NOT NULL DEFAULT 1,
    last_checked_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient queries of enabled users
CREATE INDEX IF NOT EXISTS idx_notification_users_enabled
ON notification_users(notifications_enabled) WHERE notifications_enabled = 1;

-- Push subscriptions (browser endpoints)
-- Each user can have multiple devices/browsers subscribed
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    did TEXT NOT NULL,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_secret TEXT NOT NULL,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (did) REFERENCES notification_users(did) ON DELETE CASCADE
);

-- Index for efficient lookups by DID
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_did
ON push_subscriptions(did);

-- Index for cleanup of stale subscriptions
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_last_used
ON push_subscriptions(last_used_at);

-- Notification history (prevent spam, track delivery)
CREATE TABLE IF NOT EXISTS notification_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    did TEXT NOT NULL,
    task_uri TEXT NOT NULL,
    notification_type TEXT NOT NULL CHECK(notification_type IN ('overdue', 'due_today', 'due_soon')),
    sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL CHECK(status IN ('sent', 'failed', 'expired')),
    error_message TEXT
);

-- Index for spam prevention (check if we recently notified about this task)
CREATE INDEX IF NOT EXISTS idx_notification_history_task
ON notification_history(did, task_uri, sent_at);

-- Index for cleanup of old history
CREATE INDEX IF NOT EXISTS idx_notification_history_sent_at
ON notification_history(sent_at);
