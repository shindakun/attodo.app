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
}
