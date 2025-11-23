package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shindakun/attodo/internal/models"
)

// RecurringRepo handles database operations for recurring tasks
type RecurringRepo struct {
	db *DB
}

// NewRecurringRepo creates a new recurring task repository
func NewRecurringRepo(db *DB) *RecurringRepo {
	return &RecurringRepo{db: db}
}

// ============================================================================
// RECURRING TASKS
// ============================================================================

// GetRecurringTask retrieves a recurring task by ID
func (r *RecurringRepo) GetRecurringTask(id int64) (*models.RecurringTask, error) {
	var rt models.RecurringTask
	var daysOfWeekJSON sql.NullString
	var endDate, lastGenerated, nextOccurrence sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, did, task_uri, frequency, interval, days_of_week, day_of_month,
		       end_date, max_occurrences, occurrence_count, last_generated_at,
		       next_occurrence_at, created_at, updated_at
		FROM recurring_tasks
		WHERE id = ?
	`, id).Scan(
		&rt.ID,
		&rt.DID,
		&rt.TaskURI,
		&rt.Frequency,
		&rt.Interval,
		&daysOfWeekJSON,
		&rt.DayOfMonth,
		&endDate,
		&rt.MaxOccurrences,
		&rt.OccurrenceCount,
		&lastGenerated,
		&nextOccurrence,
		&rt.CreatedAt,
		&rt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring task: %w", err)
	}

	// Handle nullable fields
	if daysOfWeekJSON.Valid {
		if err := rt.SetDaysOfWeekFromJSON(daysOfWeekJSON.String); err != nil {
			return nil, fmt.Errorf("failed to parse days of week: %w", err)
		}
	}
	if endDate.Valid {
		rt.EndDate = &endDate.Time
	}
	if lastGenerated.Valid {
		rt.LastGeneratedAt = &lastGenerated.Time
	}
	if nextOccurrence.Valid {
		rt.NextOccurrenceAt = &nextOccurrence.Time
	}

	return &rt, nil
}

// GetRecurringTaskByURI retrieves a recurring task by task URI
func (r *RecurringRepo) GetRecurringTaskByURI(taskURI string) (*models.RecurringTask, error) {
	var rt models.RecurringTask
	var daysOfWeekJSON sql.NullString
	var endDate, lastGenerated, nextOccurrence sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, did, task_uri, frequency, interval, days_of_week, day_of_month,
		       end_date, max_occurrences, occurrence_count, last_generated_at,
		       next_occurrence_at, created_at, updated_at
		FROM recurring_tasks
		WHERE task_uri = ?
	`, taskURI).Scan(
		&rt.ID,
		&rt.DID,
		&rt.TaskURI,
		&rt.Frequency,
		&rt.Interval,
		&daysOfWeekJSON,
		&rt.DayOfMonth,
		&endDate,
		&rt.MaxOccurrences,
		&rt.OccurrenceCount,
		&lastGenerated,
		&nextOccurrence,
		&rt.CreatedAt,
		&rt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring task by URI: %w", err)
	}

	// Handle nullable fields
	if daysOfWeekJSON.Valid {
		if err := rt.SetDaysOfWeekFromJSON(daysOfWeekJSON.String); err != nil {
			return nil, fmt.Errorf("failed to parse days of week: %w", err)
		}
	}
	if endDate.Valid {
		rt.EndDate = &endDate.Time
	}
	if lastGenerated.Valid {
		rt.LastGeneratedAt = &lastGenerated.Time
	}
	if nextOccurrence.Valid {
		rt.NextOccurrenceAt = &nextOccurrence.Time
	}

	return &rt, nil
}

// CreateRecurringTask creates a new recurring task
func (r *RecurringRepo) CreateRecurringTask(rt *models.RecurringTask) error {
	now := time.Now()
	rt.CreatedAt = now
	rt.UpdatedAt = now
	rt.OccurrenceCount = 0

	result, err := r.db.Exec(`
		INSERT INTO recurring_tasks (
			did, task_uri, frequency, interval, days_of_week, day_of_month,
			end_date, max_occurrences, occurrence_count, last_generated_at,
			next_occurrence_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		rt.DID,
		rt.TaskURI,
		rt.Frequency,
		rt.Interval,
		rt.DaysOfWeekJSON(),
		rt.DayOfMonth,
		rt.EndDate,
		rt.MaxOccurrences,
		rt.OccurrenceCount,
		rt.LastGeneratedAt,
		rt.NextOccurrenceAt,
		rt.CreatedAt,
		rt.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create recurring task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get recurring task ID: %w", err)
	}

	rt.ID = id
	return nil
}

// UpdateRecurringTask updates an existing recurring task
func (r *RecurringRepo) UpdateRecurringTask(rt *models.RecurringTask) error {
	rt.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		UPDATE recurring_tasks
		SET frequency = ?, interval = ?, days_of_week = ?, day_of_month = ?,
		    end_date = ?, max_occurrences = ?, occurrence_count = ?,
		    last_generated_at = ?, next_occurrence_at = ?, updated_at = ?
		WHERE id = ?
	`,
		rt.Frequency,
		rt.Interval,
		rt.DaysOfWeekJSON(),
		rt.DayOfMonth,
		rt.EndDate,
		rt.MaxOccurrences,
		rt.OccurrenceCount,
		rt.LastGeneratedAt,
		rt.NextOccurrenceAt,
		rt.UpdatedAt,
		rt.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recurring task: %w", err)
	}

	return nil
}

// DeleteRecurringTask deletes a recurring task
func (r *RecurringRepo) DeleteRecurringTask(id int64) error {
	_, err := r.db.Exec(`
		DELETE FROM recurring_tasks
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete recurring task: %w", err)
	}

	return nil
}

// GetRecurringTasksByDID retrieves all recurring tasks for a user
func (r *RecurringRepo) GetRecurringTasksByDID(did string) ([]*models.RecurringTask, error) {
	rows, err := r.db.Query(`
		SELECT id, did, task_uri, frequency, interval, days_of_week, day_of_month,
		       end_date, max_occurrences, occurrence_count, last_generated_at,
		       next_occurrence_at, created_at, updated_at
		FROM recurring_tasks
		WHERE did = ?
		ORDER BY created_at DESC
	`, did)
	if err != nil {
		return nil, fmt.Errorf("failed to query recurring tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.RecurringTask
	for rows.Next() {
		var rt models.RecurringTask
		var daysOfWeekJSON sql.NullString
		var endDate, lastGenerated, nextOccurrence sql.NullTime

		if err := rows.Scan(
			&rt.ID,
			&rt.DID,
			&rt.TaskURI,
			&rt.Frequency,
			&rt.Interval,
			&daysOfWeekJSON,
			&rt.DayOfMonth,
			&endDate,
			&rt.MaxOccurrences,
			&rt.OccurrenceCount,
			&lastGenerated,
			&nextOccurrence,
			&rt.CreatedAt,
			&rt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recurring task: %w", err)
		}

		// Handle nullable fields
		if daysOfWeekJSON.Valid {
			if err := rt.SetDaysOfWeekFromJSON(daysOfWeekJSON.String); err != nil {
				return nil, fmt.Errorf("failed to parse days of week: %w", err)
			}
		}
		if endDate.Valid {
			rt.EndDate = &endDate.Time
		}
		if lastGenerated.Valid {
			rt.LastGeneratedAt = &lastGenerated.Time
		}
		if nextOccurrence.Valid {
			rt.NextOccurrenceAt = &nextOccurrence.Time
		}

		tasks = append(tasks, &rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading recurring tasks: %w", err)
	}

	return tasks, nil
}

// GetTasksDueForGeneration retrieves recurring tasks that need new instances
func (r *RecurringRepo) GetTasksDueForGeneration() ([]*models.RecurringTask, error) {
	now := time.Now()

	rows, err := r.db.Query(`
		SELECT id, did, task_uri, frequency, interval, days_of_week, day_of_month,
		       end_date, max_occurrences, occurrence_count, last_generated_at,
		       next_occurrence_at, created_at, updated_at
		FROM recurring_tasks
		WHERE next_occurrence_at IS NOT NULL
		  AND next_occurrence_at <= ?
		  AND (end_date IS NULL OR end_date >= ?)
		  AND (max_occurrences IS NULL OR occurrence_count < max_occurrences)
		ORDER BY next_occurrence_at ASC
	`, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks due for generation: %w", err)
	}
	defer rows.Close()

	var tasks []*models.RecurringTask
	for rows.Next() {
		var rt models.RecurringTask
		var daysOfWeekJSON sql.NullString
		var endDate, lastGenerated, nextOccurrence sql.NullTime

		if err := rows.Scan(
			&rt.ID,
			&rt.DID,
			&rt.TaskURI,
			&rt.Frequency,
			&rt.Interval,
			&daysOfWeekJSON,
			&rt.DayOfMonth,
			&endDate,
			&rt.MaxOccurrences,
			&rt.OccurrenceCount,
			&lastGenerated,
			&nextOccurrence,
			&rt.CreatedAt,
			&rt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recurring task: %w", err)
		}

		// Handle nullable fields
		if daysOfWeekJSON.Valid {
			if err := rt.SetDaysOfWeekFromJSON(daysOfWeekJSON.String); err != nil {
				return nil, fmt.Errorf("failed to parse days of week: %w", err)
			}
		}
		if endDate.Valid {
			rt.EndDate = &endDate.Time
		}
		if lastGenerated.Valid {
			rt.LastGeneratedAt = &lastGenerated.Time
		}
		if nextOccurrence.Valid {
			rt.NextOccurrenceAt = &nextOccurrence.Time
		}

		tasks = append(tasks, &rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading tasks due for generation: %w", err)
	}

	return tasks, nil
}

// ============================================================================
// RECURRING INSTANCES
// ============================================================================

// CreateRecurringInstance creates a new instance record
func (r *RecurringRepo) CreateRecurringInstance(did string, recurringTaskID int64, instanceURI string, dueDate time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO recurring_instances (did, recurring_task_id, instance_task_uri, due_date, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, did, recurringTaskID, instanceURI, dueDate, time.Now())

	if err != nil {
		return fmt.Errorf("failed to create recurring instance: %w", err)
	}

	return nil
}

// MarkInstanceCompleted marks a recurring instance as completed
func (r *RecurringRepo) MarkInstanceCompleted(instanceURI string) error {
	_, err := r.db.Exec(`
		UPDATE recurring_instances
		SET completed_at = ?
		WHERE instance_task_uri = ?
	`, time.Now(), instanceURI)

	if err != nil {
		return fmt.Errorf("failed to mark instance completed: %w", err)
	}

	return nil
}

// GetInstancesByRecurringTask retrieves all instances for a recurring task
func (r *RecurringRepo) GetInstancesByRecurringTask(recurringTaskID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT instance_task_uri
		FROM recurring_instances
		WHERE recurring_task_id = ?
		ORDER BY due_date DESC
	`, recurringTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var uris []string
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, fmt.Errorf("failed to scan instance URI: %w", err)
		}
		uris = append(uris, uri)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading instances: %w", err)
	}

	return uris, nil
}
