package models

import "time"

type NotificationSettings struct {
	// Notification preferences
	NotifyOverdue bool `json:"notifyOverdue"` // Show notifications for overdue tasks
	NotifyToday   bool `json:"notifyToday"`   // Show notifications for tasks due today
	NotifySoon    bool `json:"notifySoon"`    // Show notifications for tasks due within 3 days
	HoursBefore   int  `json:"hoursBefore"`   // Hours before due date to notify (0-72)

	// Check frequency
	CheckFrequency int `json:"checkFrequency"` // Minutes between checks (15, 30, 60, 120)

	// Quiet hours
	QuietHoursEnabled bool `json:"quietHoursEnabled"` // Enable quiet hours mode
	QuietStart        int  `json:"quietStart"`        // Quiet hours start (hour 0-23)
	QuietEnd          int  `json:"quietEnd"`          // Quiet hours end (hour 0-23)

	// Notification permissions
	PushEnabled bool `json:"pushEnabled"` // Browser push notifications enabled

	// UI preferences
	TaskInputCollapsed bool `json:"taskInputCollapsed"` // Whether task input form is collapsed

	// Calendar notification settings
	CalendarNotificationsEnabled  bool              `json:"calendarNotificationsEnabled"`            // Enable calendar event notifications
	CalendarNotificationLeadTime  string            `json:"calendarNotificationLeadTime,omitempty"`  // Lead time for notifications (e.g. "1h", "30m")
	NotificationSentHistory       map[string]string `json:"notificationSentHistory,omitempty"`       // Event RKey -> last sent timestamp

	// Usage pattern tracking (for smart notification scheduling in Phase 3)
	AppUsageHours map[string]int `json:"appUsageHours,omitempty"` // Hour (0-23) -> count

	// Metadata
	UpdatedAt time.Time `json:"updatedAt"` // Last update timestamp

	// AT Protocol fields
	RKey string `json:"-"` // Record key (typically "settings")
	URI  string `json:"-"` // Full AT URI
}

// DefaultNotificationSettings returns default notification settings
func DefaultNotificationSettings() *NotificationSettings {
	return &NotificationSettings{
		NotifyOverdue:                true,
		NotifyToday:                  true,
		NotifySoon:                   false,
		HoursBefore:                  1,
		CheckFrequency:               30,
		QuietHoursEnabled:            false,
		QuietStart:                   22,
		QuietEnd:                     8,
		PushEnabled:                  false,
		TaskInputCollapsed:           false,
		CalendarNotificationsEnabled: true,
		CalendarNotificationLeadTime: "1h",
		NotificationSentHistory:      make(map[string]string),
		UpdatedAt:                    time.Now().UTC(),
		RKey:                         "settings",
	}
}
