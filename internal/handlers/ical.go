package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

// fetchEventsForDID fetches calendar events for a given DID using public read
func (h *ICalHandler) fetchEventsForDID(ctx context.Context, did string) ([]*models.CalendarEvent, error) {
	// Resolve PDS endpoint for this DID
	pds, err := h.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Build the XRPC URL for public read
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		pds, did, CalendarEventCollection)

	// Create and execute request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("XRPC error %d", resp.StatusCode)
	}

	// Parse response (reuse the same struct from calendar.go)
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

	// Convert to CalendarEvent models
	events := make([]*models.CalendarEvent, 0, len(result.Records))
	for _, record := range result.Records {
		event, err := models.ParseCalendarEvent(record.Value, record.Uri, record.Cid)
		if err != nil {
			// Skip invalid events but don't fail the whole feed
			continue
		}
		events = append(events, event)
	}

	return events, nil
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
	for _, event := range events {
		h.addEventToICalendar(&ical, event)
	}

	// iCal footer
	ical.WriteString("END:VCALENDAR\r\n")

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

// fetchTasksForDID fetches tasks for a given DID using public read
func (h *ICalHandler) fetchTasksForDID(ctx context.Context, did string) ([]*models.Task, error) {
	// Resolve PDS endpoint for this DID
	pds, err := h.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Build the XRPC URL for public read
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		pds, did, TaskCollection)

	// Create and execute request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

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

	// Convert to Task models and filter for tasks with due dates
	tasks := make([]*models.Task, 0)
	for _, record := range result.Records {
		task := parseTaskFieldsForICal(record.Value)
		task.URI = record.Uri
		task.RKey = extractRKeyForICal(record.Uri)

		// Only include tasks with due dates
		if task.DueDate != nil {
			tasks = append(tasks, &task)
		}
	}

	return tasks, nil
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
