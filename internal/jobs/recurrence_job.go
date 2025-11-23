package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/shindakun/attodo/internal/database"
)

// RecurrenceGenerationJob generates instances for recurring tasks
// This is a simple implementation - in production you might want to generate
// instances more intelligently (e.g., based on user activity)
type RecurrenceGenerationJob struct {
	repo *database.RecurringRepo
}

// NewRecurrenceGenerationJob creates a new recurrence generation job
func NewRecurrenceGenerationJob(repo *database.RecurringRepo) *RecurrenceGenerationJob {
	return &RecurrenceGenerationJob{
		repo: repo,
	}
}

// Name returns the job name
func (j *RecurrenceGenerationJob) Name() string {
	return "RecurrenceGeneration"
}

// Run executes the recurrence generation job
func (j *RecurrenceGenerationJob) Run(ctx context.Context) error {
	// Get all recurring tasks that are due for generation
	tasks, err := j.repo.GetTasksDueForGeneration()
	if err != nil {
		return fmt.Errorf("failed to get tasks due for generation: %w", err)
	}

	if len(tasks) == 0 {
		log.Println("[RecurrenceGeneration] No recurring tasks due for generation")
		return nil
	}

	log.Printf("[RecurrenceGeneration] Found %d recurring task(s) due for generation", len(tasks))

	// NOTE: This job only identifies tasks that need generation
	// The actual instance creation happens when the user completes a task
	// This is intentional - we don't want to create hundreds of future instances
	// Instead, we create the next instance on-demand when the current one is completed

	// In a more advanced implementation, you might:
	// 1. Generate instances X days/weeks in advance
	// 2. Clean up old completed instances
	// 3. Handle missed occurrences (e.g., if user was inactive)

	for _, task := range tasks {
		log.Printf("[RecurrenceGeneration] Recurring task ready: %s (next due: %v)",
			task.TaskURI, task.NextOccurrenceAt)
	}

	return nil
}
