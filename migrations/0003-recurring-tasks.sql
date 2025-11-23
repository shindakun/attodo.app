-- Recurring task tracking
-- This migration adds support for recurring tasks that regenerate after completion

-- Track recurring task patterns and metadata
-- The actual task data is stored in AT Protocol, but we track the recurrence pattern here
CREATE TABLE IF NOT EXISTS recurring_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    did TEXT NOT NULL,
    task_uri TEXT NOT NULL UNIQUE, -- AT URI of the recurring task template

    -- Recurrence pattern
    frequency TEXT NOT NULL CHECK(frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    interval INTEGER NOT NULL DEFAULT 1, -- Every N units (e.g., every 2 weeks)
    days_of_week TEXT, -- JSON array for weekly patterns: [0,1,2,3,4,5,6] where 0=Sunday
    day_of_month INTEGER, -- For monthly patterns (1-31)

    -- End conditions
    end_date DATETIME, -- Stop generating after this date
    max_occurrences INTEGER, -- Stop after N total occurrences (NULL = infinite)
    occurrence_count INTEGER NOT NULL DEFAULT 0, -- How many instances have been created

    -- Tracking
    last_generated_at DATETIME, -- When we last created an instance
    next_occurrence_at DATETIME, -- When the next instance should be created

    -- Metadata
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (did) REFERENCES notification_users(did) ON DELETE CASCADE
);

-- Index for efficient queries by user
CREATE INDEX IF NOT EXISTS idx_recurring_tasks_did
ON recurring_tasks(did);

-- Index for finding tasks that need new instances generated
CREATE INDEX IF NOT EXISTS idx_recurring_tasks_next_occurrence
ON recurring_tasks(next_occurrence_at)
WHERE next_occurrence_at IS NOT NULL;

-- Track generated instances of recurring tasks
-- Maps each generated task instance back to its recurring template
CREATE TABLE IF NOT EXISTS recurring_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    did TEXT NOT NULL,
    recurring_task_id INTEGER NOT NULL,
    instance_task_uri TEXT NOT NULL, -- AT URI of the generated task instance
    due_date DATETIME NOT NULL, -- When this instance was due
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME, -- When this instance was marked complete

    FOREIGN KEY (recurring_task_id) REFERENCES recurring_tasks(id) ON DELETE CASCADE,
    UNIQUE(recurring_task_id, instance_task_uri)
);

-- Index for finding instances by recurring task
CREATE INDEX IF NOT EXISTS idx_recurring_instances_task
ON recurring_instances(recurring_task_id);

-- Index for finding instances by user
CREATE INDEX IF NOT EXISTS idx_recurring_instances_did
ON recurring_instances(did);
