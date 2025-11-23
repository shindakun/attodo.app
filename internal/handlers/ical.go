package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/bskyoauth"
)

// ICalHandler handles iCal feed generation
type ICalHandler struct {
	client *bskyoauth.Client
}

// NewICalHandler creates a new iCal handler
func NewICalHandler(client *bskyoauth.Client) *ICalHandler {
	return &ICalHandler{
		client: client,
	}
}

// GenerateCalendarFeed generates an iCal feed for a user's calendar events
func (h *ICalHandler) GenerateCalendarFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract DID from path: /calendar/feed/{did}/events.ics
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid feed URL", http.StatusBadRequest)
		return
	}

	did := pathParts[3]
	if did == "" {
		http.Error(w, "Missing DID", http.StatusBadRequest)
		return
	}

	// Fetch events from AT Protocol (public read, no auth needed)
	events, err := h.fetchEventsForDID(ctx, did)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch events: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate iCal feed
	ical := h.generateCalendarICalendar(did, events)

	// Set headers for iCal feed
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-events.ics\"", sanitizeDID(did)))
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	// Write iCal content
	w.Write([]byte(ical))
}

// GenerateTasksFeed generates an iCal feed for a user's tasks with due dates
func (h *ICalHandler) GenerateTasksFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract DID from path: /tasks/feed/{did}/tasks.ics
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid feed URL", http.StatusBadRequest)
		return
	}

	did := pathParts[3]
	if did == "" {
		http.Error(w, "Missing DID", http.StatusBadRequest)
		return
	}

	// Fetch tasks from AT Protocol (public read, no auth needed)
	tasks, err := h.fetchTasksForDID(ctx, did)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch tasks: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate iCal feed
	ical := h.generateTasksICalendar(did, tasks)

	// Set headers for iCal feed
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-tasks.ics\"", sanitizeDID(did)))
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	// Write iCal content
	w.Write([]byte(ical))
}

// fetchEventsForDID fetches calendar events for a given DID using public read with pagination
// This includes both events owned by the DID and events they've RSVP'd to
func (h *ICalHandler) fetchEventsForDID(ctx context.Context, did string) ([]*models.CalendarEvent, error) {
	// Resolve PDS endpoint for this DID
	pds, err := h.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var allEvents []*models.CalendarEvent
	cursor := ""
	page := 0

	// Fetch own events with pagination
	log.Printf("fetchEventsForDID: Fetching own events for %s", did)
	for {
		page++
		// Build the XRPC URL for public read with optional cursor
		url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s&limit=100",
			pds, did, CalendarEventCollection)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		log.Printf("fetchEventsForDID: Fetching own events page %d (cursor: %s)", page, cursor)

		// Create and execute request
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("XRPC error %d", resp.StatusCode)
		}

		// Parse response with cursor support
		var result struct {
			Records []struct {
				Uri   string                 `json:"uri"`
				Cid   string                 `json:"cid"`
				Value map[string]interface{} `json:"value"`
			} `json:"records"`
			Cursor string `json:"cursor,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		log.Printf("fetchEventsForDID: Own events page %d returned %d records", page, len(result.Records))

		// Convert to CalendarEvent models
		for _, record := range result.Records {
			event, err := models.ParseCalendarEvent(record.Value, record.Uri, record.Cid)
			if err != nil {
				// Skip invalid events but don't fail the whole feed
				log.Printf("fetchEventsForDID: Skipping invalid event %s: %v", record.Uri, err)
				continue
			}
			log.Printf("fetchEventsForDID: Parsed event '%s' starting at %v", event.Name, event.StartsAt)
			allEvents = append(allEvents, event)
		}

		// Check if there are more pages
		if result.Cursor == "" {
			log.Printf("fetchEventsForDID: No more own event pages, total own events: %d", len(allEvents))
			break
		}
		cursor = result.Cursor
	}

	// Now fetch RSVP'd events
	log.Printf("fetchEventsForDID: Fetching RSVP'd events for %s", did)
	rsvpEvents, err := h.fetchEventsFromRSVPs(ctx, did, pds, client)
	if err != nil {
		// Don't fail if RSVP fetch fails, just log it
		log.Printf("fetchEventsForDID: Failed to fetch RSVP'd events: %v", err)
	} else {
		log.Printf("fetchEventsForDID: Found %d RSVP'd events", len(rsvpEvents))
		allEvents = append(allEvents, rsvpEvents...)
	}

	log.Printf("fetchEventsForDID: Successfully fetched %d total events (%d own + %d RSVP'd)",
		len(allEvents), len(allEvents)-len(rsvpEvents), len(rsvpEvents))
	return allEvents, nil
}

// fetchEventsFromRSVPs fetches events that the user has RSVP'd to
func (h *ICalHandler) fetchEventsFromRSVPs(ctx context.Context, did, pds string, client *http.Client) ([]*models.CalendarEvent, error) {
	var rsvpEvents []*models.CalendarEvent
	cursor := ""
	page := 0

	// Fetch all RSVPs with pagination
	for {
		page++
		url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s&limit=100",
			pds, did, CalendarRSVPCollection)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		log.Printf("fetchEventsFromRSVPs: Fetching RSVP page %d (cursor: %s)", page, cursor)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("XRPC error %d fetching RSVPs", resp.StatusCode)
		}

		var result struct {
			Records []struct {
				Uri   string                 `json:"uri"`
				Cid   string                 `json:"cid"`
				Value map[string]interface{} `json:"value"`
			} `json:"records"`
			Cursor string `json:"cursor,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		log.Printf("fetchEventsFromRSVPs: RSVP page %d returned %d records", page, len(result.Records))

		// For each RSVP, fetch the actual event
		for _, record := range result.Records {
			// Extract the event URI from the RSVP subject
			subject, ok := record.Value["subject"].(map[string]interface{})
			if !ok {
				log.Printf("fetchEventsFromRSVPs: Skipping RSVP with invalid subject")
				continue
			}

			eventURI, ok := subject["uri"].(string)
			if !ok {
				log.Printf("fetchEventsFromRSVPs: Skipping RSVP with missing event URI")
				continue
			}

			log.Printf("fetchEventsFromRSVPs: Fetching event from RSVP: %s", eventURI)

			// Fetch the event from the other user's repository
			event, err := h.fetchEventByURI(ctx, eventURI, client)
			if err != nil {
				log.Printf("fetchEventsFromRSVPs: Failed to fetch event %s: %v", eventURI, err)
				continue
			}

			rsvpEvents = append(rsvpEvents, event)
		}

		// Check if there are more pages
		if result.Cursor == "" {
			log.Printf("fetchEventsFromRSVPs: No more RSVP pages, total RSVP'd events: %d", len(rsvpEvents))
			break
		}
		cursor = result.Cursor
	}

	return rsvpEvents, nil
}

// fetchEventByURI fetches a single event by its AT URI
func (h *ICalHandler) fetchEventByURI(ctx context.Context, uri string, client *http.Client) (*models.CalendarEvent, error) {
	// Parse the URI: at://did:plc:xxx/community.lexicon.calendar.event/rkey
	if !strings.HasPrefix(uri, "at://") {
		return nil, fmt.Errorf("invalid AT URI: %s", uri)
	}

	// Extract DID and rkey from URI
	parts := strings.Split(strings.TrimPrefix(uri, "at://"), "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed AT URI: %s", uri)
	}

	eventDID := parts[0]
	collection := parts[1]
	rkey := parts[2]

	log.Printf("fetchEventByURI: Fetching event %s from %s", rkey, eventDID)

	// Resolve PDS for the event's DID
	pds, err := h.resolvePDSEndpoint(ctx, eventDID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS for %s: %w", eventDID, err)
	}

	// Fetch the event record
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pds, eventDID, collection, rkey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("XRPC error %d fetching event", resp.StatusCode)
	}

	var result struct {
		Uri   string                 `json:"uri"`
		Cid   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Parse the event
	event, err := models.ParseCalendarEvent(result.Value, result.Uri, result.Cid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	log.Printf("fetchEventByURI: Successfully fetched event '%s' starting at %v", event.Name, event.StartsAt)
	return event, nil
}

// resolvePDSEndpoint resolves the PDS endpoint for a given DID
func (h *ICalHandler) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
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

// generateCalendarICalendar generates an iCal format string from calendar events
func (h *ICalHandler) generateCalendarICalendar(did string, events []*models.CalendarEvent) string {
	log.Printf("generateCalendarICalendar: Generating iCal for %d events", len(events))
	var ical strings.Builder

	// iCal header
	ical.WriteString("BEGIN:VCALENDAR\r\n")
	ical.WriteString("VERSION:2.0\r\n")
	ical.WriteString("PRODID:-//AT Todo//Calendar Feed//EN\r\n")
	ical.WriteString(fmt.Sprintf("X-WR-CALNAME:AT Protocol Events - %s\r\n", sanitizeDID(did)))
	ical.WriteString("X-WR-TIMEZONE:UTC\r\n")
	ical.WriteString("CALSCALE:GREGORIAN\r\n")
	ical.WriteString("METHOD:PUBLISH\r\n")

	// Add each event
	for i, event := range events {
		log.Printf("generateCalendarICalendar: Adding event %d/%d: '%s'", i+1, len(events), event.Name)
		h.addEventToICalendar(&ical, event)
	}

	// iCal footer
	ical.WriteString("END:VCALENDAR\r\n")

	log.Printf("generateCalendarICalendar: iCal generation complete, size: %d bytes", ical.Len())
	return ical.String()
}

// addEventToICalendar adds a single event to the iCalendar
func (h *ICalHandler) addEventToICalendar(ical *strings.Builder, event *models.CalendarEvent) {
	ical.WriteString("BEGIN:VEVENT\r\n")

	// UID - unique identifier (use AT Protocol URI)
	ical.WriteString(fmt.Sprintf("UID:%s\r\n", event.URI))

	// DTSTAMP - when the event was created
	ical.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICalTime(event.CreatedAt)))

	// DTSTART - event start time
	if event.StartsAt != nil {
		ical.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICalTime(*event.StartsAt)))
	}

	// DTEND - event end time
	if event.EndsAt != nil {
		ical.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICalTime(*event.EndsAt)))
	}

	// SUMMARY - event title
	ical.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(event.Name)))

	// DESCRIPTION - event description
	if event.Description != "" {
		ical.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(event.Description)))
	}

	// LOCATION - event location
	if len(event.Locations) > 0 {
		location := event.Locations[0]
		if location.Name != "" {
			ical.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICalText(location.Name)))
		} else if location.Address != "" {
			ical.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICalText(location.Address)))
		}
	}

	// URL - link to Smokesignal
	if smokesignalURL := event.SmokesignalURL(); smokesignalURL != "" {
		ical.WriteString(fmt.Sprintf("URL:%s\r\n", smokesignalURL))
	}

	// STATUS - event status
	status := "CONFIRMED"
	switch event.Status {
	case models.EventStatusCancelled:
		status = "CANCELLED"
	case models.EventStatusPlanned:
		status = "TENTATIVE"
	}
	ical.WriteString(fmt.Sprintf("STATUS:%s\r\n", status))

	// CATEGORIES - attendance mode
	if event.Mode != "" {
		ical.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", strings.ToUpper(event.Mode)))
	}

	ical.WriteString("END:VEVENT\r\n")
}

// fetchTasksForDID fetches tasks for a given DID using public read with pagination
func (h *ICalHandler) fetchTasksForDID(ctx context.Context, did string) ([]*models.Task, error) {
	// Resolve PDS endpoint for this DID
	pds, err := h.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var allTasks []*models.Task
	cursor := ""
	page := 0

	// Fetch all pages
	for {
		page++
		// Build the XRPC URL for public read with optional cursor
		url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s&limit=100",
			pds, did, TaskCollection)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		log.Printf("fetchTasksForDID: Fetching page %d (cursor: %s)", page, cursor)

		// Create and execute request
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("XRPC error %d", resp.StatusCode)
		}

		// Parse response with cursor support
		var result struct {
			Records []struct {
				Uri   string                 `json:"uri"`
				Cid   string                 `json:"cid"`
				Value map[string]interface{} `json:"value"`
			} `json:"records"`
			Cursor string `json:"cursor,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		log.Printf("fetchTasksForDID: Page %d returned %d records", page, len(result.Records))

		// Convert to Task models and filter for tasks with due dates
		for _, record := range result.Records {
			task := parseTaskFieldsForICal(record.Value)
			task.URI = record.Uri
			task.RKey = extractRKeyForICal(record.Uri)

			// Only include tasks with due dates
			if task.DueDate != nil {
				log.Printf("fetchTasksForDID: Adding task '%s' with due date %v", task.Title, task.DueDate)
				allTasks = append(allTasks, &task)
			}
		}

		// Check if there are more pages
		if result.Cursor == "" {
			log.Printf("fetchTasksForDID: No more pages, total tasks with due dates: %d", len(allTasks))
			break
		}
		cursor = result.Cursor
	}

	log.Printf("fetchTasksForDID: Successfully parsed %d tasks with due dates across %d pages", len(allTasks), page)
	return allTasks, nil
}

// generateTasksICalendar generates an iCal format string from tasks
func (h *ICalHandler) generateTasksICalendar(did string, tasks []*models.Task) string {
	var ical strings.Builder

	// iCal header
	ical.WriteString("BEGIN:VCALENDAR\r\n")
	ical.WriteString("VERSION:2.0\r\n")
	ical.WriteString("PRODID:-//AT Todo//Tasks Feed//EN\r\n")
	ical.WriteString(fmt.Sprintf("X-WR-CALNAME:AT Protocol Tasks - %s\r\n", sanitizeDID(did)))
	ical.WriteString("X-WR-TIMEZONE:UTC\r\n")
	ical.WriteString("CALSCALE:GREGORIAN\r\n")
	ical.WriteString("METHOD:PUBLISH\r\n")

	// Add each task
	for _, task := range tasks {
		h.addTaskToICalendar(&ical, task)
	}

	// iCal footer
	ical.WriteString("END:VCALENDAR\r\n")

	return ical.String()
}

// addTaskToICalendar adds a single task to the iCalendar
func (h *ICalHandler) addTaskToICalendar(ical *strings.Builder, task *models.Task) {
	ical.WriteString("BEGIN:VTODO\r\n")

	// UID - unique identifier (use AT Protocol URI)
	ical.WriteString(fmt.Sprintf("UID:%s\r\n", task.URI))

	// DTSTAMP - when the task was created
	ical.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICalTime(task.CreatedAt)))

	// DUE - task due date
	if task.DueDate != nil {
		ical.WriteString(fmt.Sprintf("DUE:%s\r\n", formatICalTime(*task.DueDate)))
	}

	// SUMMARY - task title
	ical.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(task.Title)))

	// DESCRIPTION - task description
	if task.Description != "" {
		ical.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(task.Description)))
	}

	// STATUS - task completion status
	if task.Completed {
		ical.WriteString("STATUS:COMPLETED\r\n")
		if task.CompletedAt != nil {
			ical.WriteString(fmt.Sprintf("COMPLETED:%s\r\n", formatICalTime(*task.CompletedAt)))
		}
	} else {
		ical.WriteString("STATUS:NEEDS-ACTION\r\n")
	}

	// PRIORITY - map task priority (0=none, 1-10)
	// iCal priority: 1=high, 5=medium, 9=low
	priority := 5 // default medium
	ical.WriteString(fmt.Sprintf("PRIORITY:%d\r\n", priority))

	// CATEGORIES - tags
	if len(task.Tags) > 0 {
		categories := strings.Join(task.Tags, ",")
		ical.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", escapeICalText(categories)))
	}

	// URL - link to AT Todo
	// We could link to the specific task on attodo.app if we had that route
	// For now, just link to the dashboard
	ical.WriteString("URL:https://attodo.app/app\r\n")

	ical.WriteString("END:VTODO\r\n")
}

// formatICalTime formats a time.Time to iCal format (UTC)
func formatICalTime(t time.Time) string {
	// iCal format: 20060102T150405Z
	return t.UTC().Format("20060102T150405Z")
}

// escapeICalText escapes special characters in iCal text fields
func escapeICalText(text string) string {
	// Escape backslashes, commas, semicolons, and newlines
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "")
	return text
}

// sanitizeDID creates a filename-safe version of a DID
func sanitizeDID(did string) string {
	// Remove "did:plc:" prefix and use just the identifier
	parts := strings.Split(did, ":")
	if len(parts) >= 3 {
		return parts[2][:min(len(parts[2]), 16)] // Truncate to 16 chars for filename
	}
	return "calendar"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseTaskFieldsForICal parses task fields from a record map
func parseTaskFieldsForICal(record map[string]interface{}) models.Task {
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

// extractRKeyForICal extracts the record key from an AT URI
func extractRKeyForICal(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
