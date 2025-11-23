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
	"github.com/shindakun/attodo/internal/handlers"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/push"
	"github.com/shindakun/bskyoauth"
)

// CalendarNotificationJob checks for upcoming calendar events and sends notifications
type CalendarNotificationJob struct {
	repo            *database.NotificationRepo
	client          *bskyoauth.Client
	sender          *push.Sender
	calendarHandler *handlers.CalendarHandler
	settingsHandler *handlers.SettingsHandler
}

// NewCalendarNotificationJob creates a new calendar notification job
func NewCalendarNotificationJob(repo *database.NotificationRepo, client *bskyoauth.Client, sender *push.Sender, settingsHandler *handlers.SettingsHandler) *CalendarNotificationJob {
	return &CalendarNotificationJob{
		repo:            repo,
		client:          client,
		sender:          sender,
		calendarHandler: handlers.NewCalendarHandler(client),
		settingsHandler: settingsHandler,
	}
}

// Name returns the job name
func (c *CalendarNotificationJob) Name() string {
	return "CalendarNotificationCheck"
}

// Run executes the calendar notification check job
func (c *CalendarNotificationJob) Run(ctx context.Context) error {
	// Get all users with notifications enabled
	users, err := c.repo.GetEnabledNotificationUsers()
	if err != nil {
		return fmt.Errorf("failed to get enabled users: %w", err)
	}

	if len(users) == 0 {
		log.Println("[CalendarNotificationCheck] No users with notifications enabled")
		return nil
	}

	log.Printf("[CalendarNotificationCheck] Checking calendar events for %d user(s)", len(users))

	// Check events for each user
	for _, user := range users {
		if err := c.checkUserEvents(ctx, user); err != nil {
			log.Printf("[CalendarNotificationCheck] Error checking events for %s: %v", user.DID, err)
			// Continue to next user instead of failing the whole job
			continue
		}
	}

	return nil
}

// checkUserEvents checks calendar events for a single user and sends notifications
func (c *CalendarNotificationJob) checkUserEvents(ctx context.Context, user *models.NotificationUser) error {
	// Get user's push subscriptions
	subscriptions, err := c.repo.GetPushSubscriptionsByDID(user.DID)
	if err != nil {
		return fmt.Errorf("failed to get push subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		log.Printf("[CalendarNotificationCheck] User %s has no push subscriptions", user.DID)
		return nil
	}

	// Get user's calendar notification settings
	settings, err := c.getUserSettings(ctx, user.DID)
	if err != nil {
		log.Printf("[CalendarNotificationCheck] Failed to get settings for %s: %v", user.DID, err)
		// Continue with defaults
		settings = models.DefaultNotificationSettings()
	}

	// Skip if calendar notifications are disabled
	if !settings.CalendarNotificationsEnabled {
		log.Printf("[CalendarNotificationCheck] Calendar notifications disabled for %s", user.DID)
		return nil
	}

	// Get notification lead time (default 1 hour)
	leadTime := time.Hour
	if settings.CalendarNotificationLeadTime != "" {
		parsed, err := time.ParseDuration(settings.CalendarNotificationLeadTime)
		if err == nil {
			leadTime = parsed
		}
	}

	// Fetch upcoming events within the lead time window
	events, err := c.fetchUpcomingEventsForUser(ctx, user.DID, leadTime)
	if err != nil {
		return fmt.Errorf("failed to fetch upcoming events: %w", err)
	}

	if len(events) == 0 {
		return nil // No upcoming events
	}

	log.Printf("[CalendarNotificationCheck] Found %d upcoming events for %s", len(events), user.DID)

	// Send notifications for events
	for _, event := range events {
		if err := c.sendEventNotification(ctx, user.DID, event, leadTime, subscriptions); err != nil {
			log.Printf("WARNING: Failed to send notification for event %s: %v", event.RKey, err)
			continue
		}
	}

	return nil
}

// getUserSettings fetches user settings without requiring a session (public read)
func (c *CalendarNotificationJob) getUserSettings(ctx context.Context, did string) (*models.NotificationSettings, error) {
	// Resolve PDS endpoint for this DID
	pds, err := c.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Build the XRPC URL for public read (no auth needed)
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pds, did, handlers.SettingsCollection, handlers.SettingsRKey)

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
		// Settings not found is okay, return defaults
		return models.DefaultNotificationSettings(), nil
	}

	// Parse response
	var result struct {
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Parse settings
	settings := handlers.ParseSettingsRecord(result.Value)
	return settings, nil
}

// fetchUpcomingEventsForUser fetches events for a user without requiring a session (public read)
func (c *CalendarNotificationJob) fetchUpcomingEventsForUser(ctx context.Context, did string, within time.Duration) ([]*models.CalendarEvent, error) {
	// Resolve PDS endpoint for this DID
	pds, err := c.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Build the XRPC URL for public read (no auth needed)
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		pds, did, handlers.CalendarEventCollection)

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
			Cid   string                 `json:"cid"`
			Value map[string]interface{} `json:"value"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to CalendarEvent models and filter for upcoming events
	upcomingEvents := make([]*models.CalendarEvent, 0)
	for _, record := range result.Records {
		event, err := models.ParseCalendarEvent(record.Value, record.Uri, record.Cid)
		if err != nil {
			log.Printf("WARNING: Failed to parse calendar event %s: %v", record.Uri, err)
			continue
		}

		// Filter for upcoming events within the time window
		if event.StartsWithin(within) && !event.IsCancelled() {
			upcomingEvents = append(upcomingEvents, event)
		}
	}

	return upcomingEvents, nil
}

// resolvePDSEndpoint resolves the PDS endpoint for a given DID
func (c *CalendarNotificationJob) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
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

// sendEventNotification sends a notification for an event
func (c *CalendarNotificationJob) sendEventNotification(ctx context.Context, did string, event *models.CalendarEvent, leadTime time.Duration, subscriptions []*models.PushSubscription) error {
	// Check if we've already sent a notification for this event
	eventURI := event.URI
	recent, err := c.repo.GetRecentNotification(did, eventURI, 24) // Don't spam within 24 hours
	if err != nil {
		log.Printf("WARNING: Error checking notification history: %v", err)
	}
	if recent != nil {
		// Already notified recently, skip
		log.Printf("INFO: Skipping notification for event %s - already sent within 24h", event.RKey)
		return nil
	}

	// Build notification message
	title := fmt.Sprintf("Upcoming Event: %s", event.Name)
	body := c.buildNotificationBody(event, leadTime)

	notification := &push.Notification{
		Title: title,
		Body:  body,
		Icon:  "/static/icon-192.png",
		Badge: "/static/icon-192.png",
		Tag:   fmt.Sprintf("calendar-event-%s", event.RKey),
		Data: map[string]interface{}{
			"type":     "calendar_event",
			"eventURI": eventURI,
			"url":      event.SmokesignalURL(),
		},
	}

	// Send to all subscriptions
	successCount, errors := c.sender.SendToAll(subscriptions, notification)
	log.Printf("INFO: Sent calendar notification for event %s to %d/%d subscriptions", event.RKey, successCount, len(subscriptions))

	// Record notification history
	status := "sent"
	var errMsg string
	if successCount == 0 {
		status = "failed"
		if len(errors) > 0 {
			errMsg = fmt.Sprintf("%v", errors[0])
		}
	} else if len(errors) > 0 {
		errMsg = fmt.Sprintf("Sent to %d/%d subscriptions. Errors: %v", successCount, len(subscriptions), errors[0])
	}

	history := &models.NotificationHistory{
		DID:              did,
		TaskURI:          eventURI, // Reuse TaskURI field for event URI
		NotificationType: "calendar_event",
		Status:           status,
		ErrorMessage:     errMsg,
	}
	if err := c.repo.CreateNotificationHistory(history); err != nil {
		log.Printf("WARNING: Failed to create notification history: %v", err)
	}

	if successCount == 0 {
		return fmt.Errorf("failed to send to all subscriptions: %v", errors)
	}

	return nil
}

// buildNotificationBody builds the notification message body
func (c *CalendarNotificationJob) buildNotificationBody(event *models.CalendarEvent, leadTime time.Duration) string {
	if event.StartsAt == nil {
		return event.Description
	}

	timeUntil := time.Until(*event.StartsAt)

	var timeMessage string
	if timeUntil < time.Hour {
		minutes := int(timeUntil.Minutes())
		timeMessage = fmt.Sprintf("starts in %d minutes", minutes)
	} else if timeUntil < 24*time.Hour {
		hours := int(timeUntil.Hours())
		timeMessage = fmt.Sprintf("starts in %d hours", hours)
	} else {
		days := int(timeUntil.Hours() / 24)
		timeMessage = fmt.Sprintf("starts in %d days", days)
	}

	// Add mode information
	var modeInfo string
	switch event.Mode {
	case models.AttendanceModeVirtual:
		modeInfo = "💻 Virtual event"
	case models.AttendanceModeInPerson:
		modeInfo = "📍 In-person event"
	case models.AttendanceModeHybrid:
		modeInfo = "🔄 Hybrid event"
	}

	if modeInfo != "" {
		return fmt.Sprintf("%s %s", modeInfo, timeMessage)
	}

	return timeMessage
}
