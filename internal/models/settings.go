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
		NotifyOverdue:     true,
		NotifyToday:       true,
		NotifySoon:        false,
		HoursBefore:       1,
		CheckFrequency:    30,
		QuietHoursEnabled: false,
		QuietStart:        22,
		QuietEnd:          8,
		PushEnabled:       false,
		UpdatedAt:         time.Now().UTC(),
		RKey:              "settings",
	}
}
