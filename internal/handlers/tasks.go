package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/shindakun/bskyoauth"
	"github.com/shindakun/attodo/internal/dateparse"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/session"
)

const TaskCollection = "app.attodo.task"

const (
	MaxTagsPerTask = 10
	MaxTagLength   = 30
)

type TaskHandler struct {
	client      *bskyoauth.Client
	listHandler *ListHandler
}

func NewTaskHandler(client *bskyoauth.Client) *TaskHandler {
	return &TaskHandler{client: client}
}

// SetListHandler allows setting the list handler for cross-referencing
func (h *TaskHandler) SetListHandler(listHandler *ListHandler) {
	h.listHandler = listHandler
}

// withRetry executes an operation with automatic token refresh on DPoP errors
func (h *TaskHandler) withRetry(ctx context.Context, sess *bskyoauth.Session, operation func(*bskyoauth.Session) error) (*bskyoauth.Session, error) {
	var err error

	for attempt := 0; attempt < 2; attempt++ {
		err = operation(sess)
		if err == nil {
			return sess, nil
		}

		// Check if it's a DPoP replay error or 401
		if strings.Contains(err.Error(), "invalid_dpop_proof") || strings.Contains(err.Error(), "401") {
			// Refresh the token
			sess, err = h.client.RefreshToken(ctx, sess)
			if err != nil {
				return sess, err
			}
			continue
		}

		// Other errors, don't retry
		break
	}

	return sess, err
}

// getUserFriendlyError converts PDS/network errors into user-friendly messages
func getUserFriendlyError(err error, defaultMsg string) string {
	if err == nil {
		return defaultMsg
	}

	errStr := err.Error()

	// Check for specific error types
	if strings.Contains(errStr, "502") || strings.Contains(errStr, "Bad Gateway") {
		return "Your PDS server is currently unavailable (502). Please try again in a moment or contact your PDS administrator."
	}
	if strings.Contains(errStr, "503") || strings.Contains(errStr, "Service Unavailable") {
		return "Your PDS server is temporarily unavailable (503). Please try again in a moment."
	}
	if strings.Contains(errStr, "504") || strings.Contains(errStr, "Gateway Timeout") {
		return "Your PDS server timed out (504). Please try again in a moment."
	}
	if strings.Contains(errStr, "500") {
		return "Your PDS server encountered an error (500). Please contact your PDS administrator."
	}
	if strings.Contains(errStr, "EOF") || strings.Contains(errStr, "connection") {
		return "Connection to your PDS server was interrupted. Please check your network connection and try again."
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "Timeout") {
		return "Request to your PDS server timed out. Please try again."
	}

	// Return default message for other errors
	return defaultMsg
}

// parseTags parses and validates tag input from form
func parseTags(input string) []string {
	if input == "" {
		return []string{}
	}

	// Split by comma
	rawTags := strings.Split(input, ",")

	// Clean and deduplicate
	seen := make(map[string]bool)
	tags := make([]string, 0)

	for _, tag := range rawTags {
		// Trim whitespace
		cleaned := strings.TrimSpace(tag)
		if cleaned == "" {
			continue
		}

		// Enforce max length
		if len(cleaned) > MaxTagLength {
			cleaned = cleaned[:MaxTagLength]
		}

		// Deduplicate (case-insensitive)
		lower := strings.ToLower(cleaned)
		if !seen[lower] {
			seen[lower] = true
			tags = append(tags, cleaned)

			// Enforce max tags
			if len(tags) >= MaxTagsPerTask {
				break
			}
		}
	}

	return tags
}

// parseTaskFields extracts task fields from a record value map
func parseTaskFields(record map[string]interface{}) models.Task {
	task := models.Task{}

	if title, ok := record["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := record["description"].(string); ok {
		task.Description = desc
	}
	if completed, ok := record["completed"].(bool); ok {
		task.Completed = completed
	}
	if createdAt, ok := record["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			task.CreatedAt = t
		}
	}
	if completedAt, ok := record["completedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, completedAt); err == nil {
			task.CompletedAt = &t
		}
	}
	// Parse due date if present
	if dueDate, ok := record["dueDate"].(string); ok {
		if t, err := time.Parse(time.RFC3339, dueDate); err == nil {
			task.DueDate = &t
		}
	}
	// Parse tags if present
	if tags, ok := record["tags"].([]interface{}); ok {
		task.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	return task
}

// buildTaskRecord creates a task record map from a Task model
func buildTaskRecord(task *models.Task) map[string]interface{} {
	record := map[string]interface{}{
		"$type":       TaskCollection,
		"title":       task.Title,
		"description": task.Description,
		"completed":   task.Completed,
		"createdAt":   task.CreatedAt.Format(time.RFC3339),
	}

	// Add completedAt if task is completed
	if task.CompletedAt != nil {
		record["completedAt"] = task.CompletedAt.Format(time.RFC3339)
	}

	// Add due date if present
	if task.DueDate != nil {
		record["dueDate"] = task.DueDate.Format(time.RFC3339)
	}

	// Always include tags field (even if empty) to allow clearing tags
	if len(task.Tags) > 0 {
		record["tags"] = task.Tags
	} else {
		record["tags"] = []string{}
	}

	return record
}

// HandleTasks handles task CRUD operations
func (h *TaskHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListTasks(w, r)
	case http.MethodPost:
		h.handleCreateTask(w, r)
	case http.MethodPut:
		h.handleUpdateTask(w, r)
	case http.MethodPatch:
		h.handleEditTask(w, r)
	case http.MethodDelete:
		h.handleDeleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateTask creates a new task
func (h *TaskHandler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	tagsInput := r.FormValue("tags")
	dueDateInput := r.FormValue("dueDate")
	dueTimeInput := r.FormValue("dueTime")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Parse and clean tags
	tags := parseTags(tagsInput)

	// Parse date from title if no explicit due date provided
	// Use local time for parsing so "2 weeks from now" is based on user's local date
	now := time.Now()
	var dueDate *time.Time

	if dueDateInput != "" {
		// Explicit due date provided via form field
		if t, err := time.Parse("2006-01-02", dueDateInput); err == nil {
			hour, min := 0, 0
			// Parse time if provided
			if dueTimeInput != "" {
				if timeVal, err := time.Parse("15:04", dueTimeInput); err == nil {
					hour = timeVal.Hour()
					min = timeVal.Minute()
				}
			}
			// Create in local timezone, then convert to UTC
			localDate := time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, now.Location())
			dueDateUTC := localDate.UTC()
			dueDate = &dueDateUTC
		}
	} else {
		// Try to parse date and time from title using local time as reference
		parseResult := dateparse.Parse(title, now)
		if parseResult.DueDate != nil {
			// Convert to UTC properly - the parsed date is already in local timezone
			dueDateUTC := parseResult.DueDate.UTC()
			dueDate = &dueDateUTC
			// Use cleaned title (with date/time removed)
			if parseResult.CleanedTitle != "" {
				title = parseResult.CleanedTitle
			}
		}
	}

	// Create task record
	nowUTC := time.Now().UTC()
	record := map[string]interface{}{
		"$type":       TaskCollection,
		"title":       title,
		"description": description,
		"completed":   false,
		"createdAt":   nowUTC.Format(time.RFC3339),
	}

	// Add due date if parsed or provided
	if dueDate != nil {
		record["dueDate"] = dueDate.Format(time.RFC3339)
	}

	// Add tags if present
	if len(tags) > 0 {
		record["tags"] = tags
	}

	// Try to create record with retry logic
	var output *atproto.RepoCreateRecord_Output
	var err error

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		output, err = h.client.CreateRecord(r.Context(), s, TaskCollection, record)
		return err
	})

	if err != nil {
		errMsg := getUserFriendlyError(err, "Failed to create task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Extract rkey from URI
	rkey := extractRKey(output.Uri)

	// Create task model for rendering
	task := models.Task{
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   nowUTC,
		DueDate:     dueDate,
		Tags:        tags,
		RKey:        rkey,
		URI:         output.Uri,
	}

	// Return HTMX response with new task partial
	w.Header().Set("Content-Type", "text/html")
	Render(w, "task-item.html", &task)
}

// handleUpdateTask updates a task (toggle completion)
func (h *TaskHandler) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rkey := r.FormValue("rkey")
	if rkey == "" {
		http.Error(w, "rkey is required", http.StatusBadRequest)
		return
	}

	// Get the current task to toggle its completion
	var task *models.Task
	var err error

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		task, err = h.getRecord(r.Context(), s, rkey)
		return err
	})

	if err != nil {
		log.Printf("Failed to get task for update: %v", err)
		errMsg := getUserFriendlyError(err, "Failed to get task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Toggle completion
	task.Completed = !task.Completed

	// Update completedAt based on completion status
	if task.Completed {
		now := time.Now().UTC()
		task.CompletedAt = &now
	} else {
		task.CompletedAt = nil
	}

	// Build the record for update
	record := buildTaskRecord(task)

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.updateRecord(r.Context(), s, rkey, record)
	})

	if err != nil {
		log.Printf("Failed to update task after retries: %v", err)
		errMsg := getUserFriendlyError(err, "Failed to update task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Update session with new nonce after successful operation
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	log.Printf("Task updated: %s (completed: %v)", rkey, task.Completed)

	// Return empty response to trigger deletion from current view
	// The task will appear in the other tab when reloaded
	w.WriteHeader(http.StatusOK)
}

// handleEditTask edits task title and description
func (h *TaskHandler) handleEditTask(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rkey := r.FormValue("rkey")
	title := r.FormValue("title")

	if rkey == "" {
		http.Error(w, "rkey is required", http.StatusBadRequest)
		return
	}

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Get the current task
	var task *models.Task
	var err error

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		task, err = h.getRecord(r.Context(), s, rkey)
		return err
	})

	if err != nil {
		log.Printf("Failed to get task for edit: %v", err)
		errMsg := getUserFriendlyError(err, "Failed to get task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Update title and description
	task.Title = title
	task.Description = r.FormValue("description")

	// Update tags
	tagsInput := r.FormValue("tags")
	task.Tags = parseTags(tagsInput)

	// Update due date and time
	dueDateInput := r.FormValue("dueDate")
	dueTimeInput := r.FormValue("dueTime")

	if dueDateInput != "" {
		// Explicit due date provided
		if t, err := time.Parse("2006-01-02", dueDateInput); err == nil {
			hour, min := 0, 0
			// Parse time if provided
			if dueTimeInput != "" {
				if timeVal, err := time.Parse("15:04", dueTimeInput); err == nil {
					hour = timeVal.Hour()
					min = timeVal.Minute()
				}
			}
			// Create in local timezone, then convert to UTC
			now := time.Now()
			localDate := time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, now.Location())
			dueDateUTC := localDate.UTC()
			task.DueDate = &dueDateUTC
		}
	} else {
		// No explicit date - try parsing from title using local time as reference
		parseResult := dateparse.Parse(task.Title, time.Now())
		if parseResult.DueDate != nil {
			// Convert to UTC properly - the parsed date is already in local timezone
			dueDateUTC := parseResult.DueDate.UTC()
			task.DueDate = &dueDateUTC
			// Use cleaned title
			if parseResult.CleanedTitle != "" {
				task.Title = parseResult.CleanedTitle
			}
		} else {
			// No date in title either - clear due date
			task.DueDate = nil
		}
	}

	// Build the record for update
	record := buildTaskRecord(task)

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.updateRecord(r.Context(), s, rkey, record)
	})

	if err != nil {
		log.Printf("Failed to edit task after retries: %v", err)
		errMsg := getUserFriendlyError(err, "Failed to edit task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Update session with new nonce after successful operation
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	log.Printf("Task edited: %s", rkey)

	// Return updated task partial for HTMX to swap
	w.Header().Set("Content-Type", "text/html")
	Render(w, "task-item.html", task) // task is already a pointer from getRecord
}

// handleDeleteTask deletes a task
func (h *TaskHandler) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rkey from URL or form
	rkey := r.URL.Query().Get("rkey")
	if rkey == "" {
		rkey = r.FormValue("rkey")
	}

	if rkey == "" {
		http.Error(w, "rkey is required", http.StatusBadRequest)
		return
	}

	// Try to delete record with retry logic
	sess, err := h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.client.DeleteRecord(r.Context(), s, TaskCollection, rkey)
	})

	if err != nil {
		log.Printf("Failed to delete task after retries: %v", err)
		errMsg := getUserFriendlyError(err, "Failed to delete task. Please try again.")
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	log.Printf("Task deleted: %s for DID: %s", rkey, sess.DID)

	// Return empty response for HTMX to remove element
	w.WriteHeader(http.StatusOK)
}

// handleListTasks lists all tasks for the user
func (h *TaskHandler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get filter parameters
	filter := r.URL.Query().Get("filter")
	tagFilter := r.URL.Query().Get("tag")
	sortBy := r.URL.Query().Get("sort")
	dueFilter := r.URL.Query().Get("due")

	log.Printf("Listing tasks for DID: %s (filter: %s, tag: %s, sort: %s, due: %s)", sess.DID, filter, tagFilter, sortBy, dueFilter)

	// Use com.atproto.repo.listRecords to fetch all tasks
	var tasks []models.Task
	var err error

	sess, err = h.withRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		tasks, err = h.listRecords(r.Context(), s)
		return err
	})

	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		// Return empty list on error rather than failing
		tasks = []models.Task{}
	}

	// Fetch all lists to populate task-to-list relationships
	if h.listHandler != nil {
		var lists []*models.TaskList
		sess, err = h.listHandler.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
			lists, err = h.listHandler.ListRecords(r.Context(), s)
			return err
		})

		if err == nil && lists != nil {
			// Create a map of task URI to lists
			taskListMap := make(map[string][]*models.TaskList)
			for _, list := range lists {
				for _, taskURI := range list.TaskURIs {
					taskListMap[taskURI] = append(taskListMap[taskURI], list)
				}
			}

			// Populate the Lists field for each task
			for i := range tasks {
				taskURI := tasks[i].URI
				if taskLists, exists := taskListMap[taskURI]; exists {
					tasks[i].Lists = taskLists
				}
			}
		}
	}

	// Filter tasks based on completion status and tags
	filteredTasks := make([]models.Task, 0)
	for _, task := range tasks {
		// Apply completion filter
		if filter == "completed" && !task.Completed {
			continue
		} else if filter == "incomplete" && task.Completed {
			continue
		}

		// Apply tag filter
		if tagFilter != "" {
			hasTag := false
			for _, tag := range task.Tags {
				if strings.EqualFold(tag, tagFilter) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Apply due date filter
		if dueFilter != "" {
			switch dueFilter {
			case "overdue":
				if !task.IsOverdue() {
					continue
				}
			case "today":
				if !task.IsDueToday() {
					continue
				}
			case "upcoming":
				if !task.IsDueSoon() {
					continue
				}
			case "none":
				if task.DueDate != nil {
					continue
				}
			case "has":
				// Show only tasks that have a due date
				if task.DueDate == nil {
					continue
				}
			}
		}

		filteredTasks = append(filteredTasks, task)
	}

	// Sort tasks
	sortTasks(filteredTasks, sortBy)

	log.Printf("Found %d tasks (filtered: %d)", len(tasks), len(filteredTasks))

	// Check if client wants JSON response
	acceptHeader := r.Header.Get("Accept")
	formatParam := r.URL.Query().Get("format")

	if formatParam == "json" || strings.Contains(acceptHeader, "application/json") {
		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredTasks); err != nil {
			log.Printf("Failed to encode JSON: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Return HTML partials for HTMX
	w.Header().Set("Content-Type", "text/html")
	for i := range filteredTasks {
		if err := Render(w, "task-item.html", &filteredTasks[i]); err != nil {
			log.Printf("Failed to render task: %v", err)
		}
	}
}

// sortTasks sorts tasks by different criteria
func sortTasks(tasks []models.Task, sortBy string) {
	switch sortBy {
	case "due":
		// Sort by due date: tasks with due dates first (soonest first), then tasks without due dates
		sort.Slice(tasks, func(i, j int) bool {
			// Tasks without due dates go last
			if tasks[i].DueDate == nil && tasks[j].DueDate == nil {
				return false // Keep original order
			}
			if tasks[i].DueDate == nil {
				return false // i goes after j
			}
			if tasks[j].DueDate == nil {
				return true // i goes before j
			}
			// Both have due dates - sort by date (earliest first)
			return tasks[i].DueDate.Before(*tasks[j].DueDate)
		})
	case "title":
		// Sort alphabetically by title (case-insensitive)
		sort.Slice(tasks, func(i, j int) bool {
			return strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
		})
	case "created":
		// Sort by creation date (newest first)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	default:
		// Default: most recent first (by creation date)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	}
}

// listRecords fetches all records from a collection using com.atproto.repo.listRecords
func (h *TaskHandler) listRecords(ctx context.Context, sess *bskyoauth.Session) ([]models.Task, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		sess.PDS, sess.DID, TaskCollection)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Records []struct {
			Uri   string                 `json:"uri"`
			Value map[string]interface{} `json:"value"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to Task models
	tasks := make([]models.Task, 0, len(result.Records))
	for _, record := range result.Records {
		task := parseTaskFields(record.Value)
		task.URI = record.Uri
		task.RKey = extractRKey(record.Uri)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// getRecord fetches a single record using direct XRPC call (same as listRecords does)
func (h *TaskHandler) getRecord(ctx context.Context, sess *bskyoauth.Session, rkey string) (*models.Task, error) {
	// Build the XRPC URL (same pattern as listRecords uses)
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, sess.DID, TaskCollection, rkey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Use Bearer token for read operations (same as listRecords)
	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (same structure as listRecords)
	var result struct {
		Uri   string                 `json:"uri"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	task := parseTaskFields(result.Value)
	task.URI = result.Uri
	task.RKey = rkey

	return &task, nil
}

// updateRecord updates a record by making a direct HTTP request to the PDS
func (h *TaskHandler) updateRecord(ctx context.Context, sess *bskyoauth.Session, rkey string, record map[string]interface{}) error {
	log.Printf("updateRecord: DID=%s, Collection=%s, RKey=%s", sess.DID, TaskCollection, rkey)

	// Resolve the actual PDS endpoint for this user (same as CreateRecord does)
	pdsHost, err := h.resolvePDSEndpoint(ctx, sess.DID)
	if err != nil {
		return fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}
	log.Printf("updateRecord: Resolved PDS=%s", pdsHost)

	// Add $type field to the record if not present
	if _, exists := record["$type"]; !exists {
		record["$type"] = TaskCollection
	}

	// Build the request body
	body := map[string]interface{}{
		"repo":       sess.DID,
		"collection": TaskCollection,
		"rkey":       rkey,
		"record":     record,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request to the resolved PDS endpoint
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.putRecord", pdsHost)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Create DPoP transport for authentication
	dpopTransport := bskyoauth.NewDPoPTransport(
		http.DefaultTransport,
		sess.DPoPKey,
		sess.AccessToken,
		sess.DPoPNonce,
	)

	httpClient := &http.Client{
		Transport: dpopTransport,
		Timeout:   10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("updateRecord: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var output atproto.RepoPutRecord_Output
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("updateRecord: Success! URI=%s", output.Uri)
	return nil
}

// resolvePDSEndpoint resolves the PDS endpoint for a given DID (same as bskyoauth internal API does)
func (h *TaskHandler) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
	dir := identity.DefaultDirectory()
	atid, err := syntax.ParseAtIdentifier(did)
	if err != nil {
		return "", err
	}

	ident, err := dir.Lookup(ctx, *atid)
	if err != nil {
		return "", err
	}

	return ident.PDSEndpoint(), nil
}

// extractRKey extracts the record key from an AT URI
func extractRKey(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
