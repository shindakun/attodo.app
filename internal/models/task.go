package models

import "time"

// Task represents a todo item stored in AT Protocol
type Task struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"` // Pointer so it can be nil/omitted
	DueDate     *time.Time `json:"dueDate,omitempty"`     // Due date for the task
	Tags        []string   `json:"tags,omitempty"`        // User-defined tags for categorization

	// Metadata from AT Protocol (populated after creation)
	RKey string `json:"-"` // Record key (extracted from URI)
	URI  string `json:"-"` // Full AT URI

	// Transient field - populated when fetching task with list memberships
	Lists []*TaskList `json:"-"` // Lists this task belongs to (not stored in AT Protocol)
}

// TaskList represents a collection of tasks stored in AT Protocol
type TaskList struct {
	Name        string    `json:"name"`                  // Name of the list (e.g., "Work", "Personal", "Shopping")
	Description string    `json:"description,omitempty"` // Optional description of the list
	TaskURIs    []string  `json:"taskUris"`              // Array of AT URIs referencing tasks
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Metadata from AT Protocol (populated after creation)
	RKey        string `json:"-"` // Record key (extracted from URI)
	URI         string `json:"-"` // Full AT URI
	OwnerHandle string `json:"-"` // Handle of the list owner (for public views)

	// Transient field - populated when fetching list with tasks
	Tasks []*Task `json:"-"` // Resolved task objects (not stored in AT Protocol)
}

// IsOverdue returns true if task has a due date in the past and is not completed
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Completed {
		return false
	}
	// Use local timezone for comparison
	now := time.Now()
	due := t.DueDate.In(now.Location())

	// Set time to start of day for fair comparison
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return dueDay.Before(today)
}

// IsDueToday returns true if task is due today
func (t *Task) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	// Compare in local timezone, not UTC
	now := time.Now()
	due := t.DueDate.In(now.Location())
	return now.Year() == due.Year() &&
		now.Month() == due.Month() &&
		now.Day() == due.Day()
}

// IsDueSoon returns true if task is due within next 3 days (not including today)
func (t *Task) IsDueSoon() bool {
	if t.DueDate == nil || t.Completed {
		return false
	}
	// Use local timezone for comparison
	now := time.Now()
	due := t.DueDate.In(now.Location())

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	threeDaysFromNow := today.AddDate(0, 0, 3)
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())

	// Due soon means: after today and before/on 3 days from now
	return dueDay.After(today) && (dueDay.Before(threeDaysFromNow) || dueDay.Equal(threeDaysFromNow))
}

// DueDateDisplay returns a human-friendly due date string
func (t *Task) DueDateDisplay() string {
	if t.DueDate == nil {
		return ""
	}

	// Use local timezone for display, not UTC
	now := time.Now()
	due := t.DueDate.In(now.Location())

	// Check if time is set (not midnight in LOCAL timezone)
	// We check local time because a time like 4pm PST = 00:00 UTC (midnight)
	hasTime := due.Hour() != 0 || due.Minute() != 0
	timeStr := ""
	if hasTime {
		timeStr = " at " + due.Format("3:04pm")
	}

	// Today
	if t.IsDueToday() {
		return "Today" + timeStr
	}

	// Tomorrow
	tomorrow := now.AddDate(0, 0, 1)
	if tomorrow.Year() == due.Year() &&
		tomorrow.Month() == due.Month() &&
		tomorrow.Day() == due.Day() {
		return "Tomorrow" + timeStr
	}

	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if yesterday.Year() == due.Year() &&
		yesterday.Month() == due.Month() &&
		yesterday.Day() == due.Day() {
		return "Yesterday" + timeStr
	}

	// This week (show day name)
	daysUntil := int(due.Sub(now).Hours() / 24)
	if daysUntil > 0 && daysUntil < 7 {
		return due.Format("Monday") + timeStr
	}

	// This year (omit year)
	if due.Year() == now.Year() {
		return due.Format("Jan 2") + timeStr
	}

	// Other years
	return due.Format("Jan 2, 2006") + timeStr
}
