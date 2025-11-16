package models

import "time"

// Task represents a todo item stored in AT Protocol
type Task struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"` // Pointer so it can be nil/omitted

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
	RKey string `json:"-"` // Record key (extracted from URI)
	URI  string `json:"-"` // Full AT URI

	// Transient field - populated when fetching list with tasks
	Tasks []*Task `json:"-"` // Resolved task objects (not stored in AT Protocol)
}
