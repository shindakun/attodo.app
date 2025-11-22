package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/shindakun/attodo/internal/database"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/push"
	"github.com/shindakun/bskyoauth"
)

const (
	NOTIFICATION_COOLDOWN_HOURS = 12 // Don't spam the same task within 12 hours
)

// NotificationCheckJob checks for due tasks and sends push notifications
type NotificationCheckJob struct {
	repo   *database.NotificationRepo
	client *bskyoauth.Client
	sender *push.Sender
}

// NewNotificationCheckJob creates a new notification check job
func NewNotificationCheckJob(repo *database.NotificationRepo, client *bskyoauth.Client, sender *push.Sender) *NotificationCheckJob {
	return &NotificationCheckJob{
		repo:   repo,
		client: client,
		sender: sender,
	}
}

// Name returns the job name
func (j *NotificationCheckJob) Name() string {
	return "NotificationCheck"
}

// Run executes the notification check job
func (j *NotificationCheckJob) Run(ctx context.Context) error {
	// Get all users with notifications enabled
	users, err := j.repo.GetEnabledNotificationUsers()
	if err != nil {
		return fmt.Errorf("failed to get enabled users: %w", err)
	}

	if len(users) == 0 {
		log.Println("[NotificationCheck] No users with notifications enabled")
		return nil
	}

	log.Printf("[NotificationCheck] Checking tasks for %d user(s)", len(users))

	// Check tasks for each user
	for _, user := range users {
		if err := j.checkUserTasks(ctx, user); err != nil {
			log.Printf("[NotificationCheck] Error checking tasks for %s: %v", user.DID, err)
			// Continue to next user instead of failing the whole job
			continue
		}

		// Update last checked time
		user.LastCheckedAt = timePtr(time.Now())
		if err := j.repo.UpdateNotificationUser(user); err != nil {
			log.Printf("[NotificationCheck] Failed to update last checked time for %s: %v", user.DID, err)
		}
	}

	return nil
}

// checkUserTasks checks tasks for a single user and sends notifications
func (j *NotificationCheckJob) checkUserTasks(ctx context.Context, user *models.NotificationUser) error {
	// Get user's push subscriptions
	subscriptions, err := j.repo.GetPushSubscriptionsByDID(user.DID)
	if err != nil {
		return fmt.Errorf("failed to get push subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		log.Printf("[NotificationCheck] User %s has no push subscriptions", user.DID)
		return nil
	}

	// Fetch user's tasks from AT Protocol
	tasks, err := j.fetchUserTasks(ctx, user.DID)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil // No tasks, nothing to do
	}

	// Group tasks by notification type
	overdue := make([]*models.Task, 0)
	dueToday := make([]*models.Task, 0)
	dueSoon := make([]*models.Task, 0)

	for _, task := range tasks {
		if task.Completed || task.DueDate == nil {
			continue
		}

		// Check if we recently notified about this task
		recent, err := j.repo.GetRecentNotification(user.DID, task.URI, NOTIFICATION_COOLDOWN_HOURS)
		if err != nil {
			log.Printf("[NotificationCheck] Error checking notification history: %v", err)
			continue
		}
		if recent != nil {
			// Already notified recently, skip
			continue
		}

		if task.IsOverdue() {
			overdue = append(overdue, task)
		} else if task.IsDueToday() {
			dueToday = append(dueToday, task)
		} else if task.IsDueSoon() {
			dueSoon = append(dueSoon, task)
		}
	}

	// Send notifications (prioritize overdue > today > soon)
	if len(overdue) > 0 {
		return j.sendOverdueNotification(user.DID, overdue, subscriptions)
	}
	if len(dueToday) > 0 {
		return j.sendDueTodayNotification(user.DID, dueToday, subscriptions)
	}
	if len(dueSoon) > 0 {
		return j.sendDueSoonNotification(user.DID, dueSoon, subscriptions)
	}

	return nil
}

// fetchUserTasks fetches incomplete tasks for a user from AT Protocol
func (j *NotificationCheckJob) fetchUserTasks(ctx context.Context, did string) ([]*models.Task, error) {
	// Resolve PDS endpoint for this DID
	pds, err := j.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Build the XRPC URL for public read (no auth needed)
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=app.attodo.task",
		pds, did)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Make request (no auth needed for public reads)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("XRPC error %d", resp.StatusCode)
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

	// Convert to Task models and filter for incomplete tasks
	tasks := make([]*models.Task, 0)
	for _, record := range result.Records {
		task := j.parseTaskFields(record.Value)
		task.URI = record.Uri

		// Only include incomplete tasks
		if !task.Completed {
			tasks = append(tasks, task)
		}
	}

	log.Printf("[NotificationCheck] Fetched %d incomplete tasks for %s", len(tasks), did)
	return tasks, nil
}

// sendOverdueNotification sends a notification for overdue tasks
func (j *NotificationCheckJob) sendOverdueNotification(did string, tasks []*models.Task, subs []*models.PushSubscription) error {
	title := fmt.Sprintf("%d Overdue Task%s", len(tasks), pluralize(len(tasks)))
	body := buildTaskList(tasks, 3)

	notification := &push.Notification{
		Title: title,
		Body:  body,
		Icon:  "/static/icon-192.png",
		Badge: "/static/icon-192.png",
		Tag:   "overdue-tasks",
		Data: map[string]interface{}{
			"type":  "overdue",
			"count": len(tasks),
		},
	}

	successCount, errors := j.sender.SendToAll(subs, notification)
	log.Printf("[NotificationCheck] Sent overdue notification to %d/%d subscriptions", successCount, len(subs))

	// Record notification history for each task
	// Mark as "sent" if at least one subscription succeeded
	for _, task := range tasks {
		status := "sent"
		var errMsg string
		if successCount == 0 {
			// Only mark as failed if ALL subscriptions failed
			status = "failed"
			if len(errors) > 0 {
				errMsg = fmt.Sprintf("%v", errors[0])
			}
		} else if len(errors) > 0 {
			// Partial success - note the errors but mark as sent
			errMsg = fmt.Sprintf("Sent to %d/%d subscriptions. Errors: %v", successCount, len(subs), errors[0])
		}

		history := &models.NotificationHistory{
			DID:              did,
			TaskURI:          task.URI,
			NotificationType: "overdue",
			Status:           status,
			ErrorMessage:     errMsg,
		}
		if err := j.repo.CreateNotificationHistory(history); err != nil {
			log.Printf("[NotificationCheck] Failed to create notification history: %v", err)
		}
	}

	if successCount == 0 {
		// Only return error if ALL subscriptions failed
		return fmt.Errorf("failed to send to all subscriptions: %v", errors)
	}
	if len(errors) > 0 {
		log.Printf("[NotificationCheck] Partial success: %d/%d subscriptions succeeded", successCount, len(subs))
	}
	return nil
}

// sendDueTodayNotification sends a notification for tasks due today
func (j *NotificationCheckJob) sendDueTodayNotification(did string, tasks []*models.Task, subs []*models.PushSubscription) error {
	title := fmt.Sprintf("%d Task%s Due Today", len(tasks), pluralize(len(tasks)))
	body := buildTaskList(tasks, 3)

	notification := &push.Notification{
		Title: title,
		Body:  body,
		Icon:  "/static/icon-192.png",
		Badge: "/static/icon-192.png",
		Tag:   "due-today",
		Data: map[string]interface{}{
			"type":  "due_today",
			"count": len(tasks),
		},
	}

	successCount, errors := j.sender.SendToAll(subs, notification)
	log.Printf("[NotificationCheck] Sent due today notification to %d/%d subscriptions", successCount, len(subs))

	// Record notification history
	// Mark as "sent" if at least one subscription succeeded
	for _, task := range tasks {
		status := "sent"
		var errMsg string
		if successCount == 0 {
			// Only mark as failed if ALL subscriptions failed
			status = "failed"
			if len(errors) > 0 {
				errMsg = fmt.Sprintf("%v", errors[0])
			}
		} else if len(errors) > 0 {
			// Partial success - note the errors but mark as sent
			errMsg = fmt.Sprintf("Sent to %d/%d subscriptions. Errors: %v", successCount, len(subs), errors[0])
		}

		history := &models.NotificationHistory{
			DID:              did,
			TaskURI:          task.URI,
			NotificationType: "due_today",
			Status:           status,
			ErrorMessage:     errMsg,
		}
		if err := j.repo.CreateNotificationHistory(history); err != nil {
			log.Printf("[NotificationCheck] Failed to create notification history: %v", err)
		}
	}

	if successCount == 0 {
		// Only return error if ALL subscriptions failed
		return fmt.Errorf("failed to send to all subscriptions: %v", errors)
	}
	if len(errors) > 0 {
		log.Printf("[NotificationCheck] Partial success: %d/%d subscriptions succeeded", successCount, len(subs))
	}
	return nil
}

// sendDueSoonNotification sends a notification for tasks due soon
func (j *NotificationCheckJob) sendDueSoonNotification(did string, tasks []*models.Task, subs []*models.PushSubscription) error {
	title := fmt.Sprintf("%d Task%s Due Soon", len(tasks), pluralize(len(tasks)))
	body := buildTaskList(tasks, 3)

	notification := &push.Notification{
		Title: title,
		Body:  body,
		Icon:  "/static/icon-192.png",
		Badge: "/static/icon-192.png",
		Tag:   "due-soon",
		Data: map[string]interface{}{
			"type":  "due_soon",
			"count": len(tasks),
		},
	}

	successCount, errors := j.sender.SendToAll(subs, notification)
	log.Printf("[NotificationCheck] Sent due soon notification to %d/%d subscriptions", successCount, len(subs))

	// Record notification history
	// Mark as "sent" if at least one subscription succeeded
	for _, task := range tasks {
		status := "sent"
		var errMsg string
		if successCount == 0 {
			// Only mark as failed if ALL subscriptions failed
			status = "failed"
			if len(errors) > 0 {
				errMsg = fmt.Sprintf("%v", errors[0])
			}
		} else if len(errors) > 0 {
			// Partial success - note the errors but mark as sent
			errMsg = fmt.Sprintf("Sent to %d/%d subscriptions. Errors: %v", successCount, len(subs), errors[0])
		}

		history := &models.NotificationHistory{
			DID:              did,
			TaskURI:          task.URI,
			NotificationType: "due_soon",
			Status:           status,
			ErrorMessage:     errMsg,
		}
		if err := j.repo.CreateNotificationHistory(history); err != nil {
			log.Printf("[NotificationCheck] Failed to create notification history: %v", err)
		}
	}

	if successCount == 0 {
		// Only return error if ALL subscriptions failed
		return fmt.Errorf("failed to send to all subscriptions: %v", errors)
	}
	if len(errors) > 0 {
		log.Printf("[NotificationCheck] Partial success: %d/%d subscriptions succeeded", successCount, len(subs))
	}
	return nil
}

// Helper functions

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func buildTaskList(tasks []*models.Task, limit int) string {
	body := ""
	for i, task := range tasks {
		if i >= limit {
			more := len(tasks) - limit
			body += fmt.Sprintf("\n...and %d more", more)
			break
		}
		body += fmt.Sprintf("• %s\n", task.Title)
	}
	return body
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// resolvePDSEndpoint resolves the PDS endpoint for a given DID
func (j *NotificationCheckJob) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
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

// parseTaskFields extracts task fields from a record value map
func (j *NotificationCheckJob) parseTaskFields(record map[string]interface{}) *models.Task {
	task := &models.Task{}

	if title, ok := record["title"].(string); ok {
		task.Title = title
	}
	if description, ok := record["description"].(string); ok {
		task.Description = description
	}
	if completed, ok := record["completed"].(bool); ok {
		task.Completed = completed
	}

	// Parse timestamps
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
	if dueDate, ok := record["dueDate"].(string); ok {
		if t, err := time.Parse(time.RFC3339, dueDate); err == nil {
			task.DueDate = &t
		}
	}

	// Parse tags
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
