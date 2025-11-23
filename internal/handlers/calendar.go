package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/session"
	"github.com/shindakun/bskyoauth"
)

const (
	CalendarEventCollection = "community.lexicon.calendar.event"
	CalendarRSVPCollection  = "community.lexicon.calendar.rsvp"
)

type CalendarHandler struct {
	client *bskyoauth.Client
}

func NewCalendarHandler(client *bskyoauth.Client) *CalendarHandler {
	return &CalendarHandler{client: client}
}

// withRetry executes an operation with automatic token refresh on DPoP errors
func (h *CalendarHandler) withRetry(ctx context.Context, sess *bskyoauth.Session, operation func(*bskyoauth.Session) error) (*bskyoauth.Session, error) {
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

// ListEvents fetches all calendar events (both owned and RSVP'd)
func (h *CalendarHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var events []*models.CalendarEvent
	var err error

	sess, err = h.withRetry(ctx, sess, func(s *bskyoauth.Session) error {
		// Fetch events from user's own repository
		ownEvents, err := h.ListEventRecords(ctx, s)
		if err != nil {
			fmt.Printf("WARNING: Failed to fetch own events: %v\n", err)
		} else {
			events = append(events, ownEvents...)
		}

		// Fetch events user has RSVP'd to
		rsvpEvents, err := h.ListEventsFromRSVPs(ctx, s)
		if err != nil {
			fmt.Printf("WARNING: Failed to fetch RSVP'd events: %v\n", err)
		} else {
			events = append(events, rsvpEvents...)
		}

		return nil
	})

	if err != nil {
		http.Error(w, getUserFriendlyError(err, "Failed to fetch calendar events"), http.StatusInternalServerError)
		return
	}

	// Sort events by StartsAt in reverse chronological order (newest first)
	sortEventsByDate(events)

	// Check if client wants HTML or JSON based on Accept header
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/html") || r.Header.Get("HX-Request") == "true" {
		// Return HTML for HTMX
		if len(events) == 0 {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<p style=\"color: var(--pico-muted-color); text-align: center; padding: 2rem;\">No calendar events found. Events created in other AT Protocol calendar apps will appear here.</p>"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		for _, event := range events {
			Render(w, "calendar-event-card.html", event)
		}
		return
	}

	// Return JSON for API calls
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetEvent fetches a single calendar event by rkey
func (h *CalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rkey from URL path
	rkey := strings.TrimPrefix(r.URL.Path, "/app/calendar/events/")
	if idx := strings.Index(rkey, "/"); idx != -1 {
		rkey = rkey[:idx]
	}

	if rkey == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	var event *models.CalendarEvent
	var err error

	// Try to fetch from user's own repository first
	sess, err = h.withRetry(ctx, sess, func(s *bskyoauth.Session) error {
		event, err = h.getEventRecord(ctx, s, rkey)
		return err
	})

	// If not found in user's repository, search through all events (including RSVP'd)
	if err != nil && (strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "RecordNotFound")) {
		fmt.Printf("DEBUG: Event %s not found in user's own repository, searching RSVP'd events...\n", rkey)
		sess, err = h.withRetry(ctx, sess, func(s *bskyoauth.Session) error {
			// Get all events (own + RSVP'd)
			var allEvents []*models.CalendarEvent

			ownEvents, err := h.ListEventRecords(ctx, s)
			if err == nil {
				allEvents = append(allEvents, ownEvents...)
			}

			rsvpEvents, err := h.ListEventsFromRSVPs(ctx, s)
			if err == nil {
				allEvents = append(allEvents, rsvpEvents...)
			}

			fmt.Printf("DEBUG: Searching through %d total events for rkey %s\n", len(allEvents), rkey)

			// Find event with matching rkey
			for _, e := range allEvents {
				if e.RKey == rkey {
					fmt.Printf("DEBUG: Found event %s in all events list\n", rkey)
					event = e
					return nil
				}
			}

			return fmt.Errorf("event not found: %s", rkey)
		})
	}

	if err != nil {
		fmt.Printf("ERROR: Failed to fetch event %s: %v\n", rkey, err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "RecordNotFound") {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, getUserFriendlyError(err, "Failed to fetch event"), http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: Successfully fetched event %s\n", rkey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// GetEventRSVPs fetches RSVPs for a specific event
func (h *CalendarHandler) GetEventRSVPs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rkey from URL path
	path := strings.TrimPrefix(r.URL.Path, "/app/calendar/events/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	rkey := parts[0]

	// Construct the event URI
	eventURI := fmt.Sprintf("at://%s/%s/%s", sess.DID, CalendarEventCollection, rkey)

	var rsvps []*models.CalendarRSVP
	var err error

	sess, err = h.withRetry(ctx, sess, func(s *bskyoauth.Session) error {
		allRSVPs, err := h.listRSVPRecords(ctx, s)
		if err != nil {
			return err
		}

		// Filter for this event
		for _, rsvp := range allRSVPs {
			if rsvp.Subject.URI == eventURI {
				rsvps = append(rsvps, rsvp)
			}
		}

		return nil
	})

	if err != nil {
		http.Error(w, getUserFriendlyError(err, "Failed to fetch RSVPs"), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsvps)
}

// ListUpcomingEvents fetches events starting within a specified time window
func (h *CalendarHandler) ListUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse duration from query params (default: 7 days)
	durationStr := r.URL.Query().Get("within")
	duration := 7 * 24 * time.Hour // default 7 days
	if durationStr != "" {
		if parsed, err := time.ParseDuration(durationStr); err == nil {
			duration = parsed
		}
	}

	var events []*models.CalendarEvent
	var err error

	sess, err = h.withRetry(ctx, sess, func(s *bskyoauth.Session) error {
		allEvents, err := h.ListEventRecords(ctx, s)
		if err != nil {
			return err
		}

		// Filter for upcoming events
		for _, event := range allEvents {
			if event.IsUpcoming() && event.StartsWithin(duration) && !event.IsCancelled() {
				events = append(events, event)
			}
		}

		return nil
	})

	if err != nil {
		http.Error(w, getUserFriendlyError(err, "Failed to fetch upcoming events"), http.StatusInternalServerError)
		return
	}

	// Check if client wants HTML or JSON based on Accept header
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/html") || r.Header.Get("HX-Request") == "true" {
		// Return HTML for HTMX
		if len(events) == 0 {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<p style=\"color: var(--pico-muted-color); text-align: center; padding: 2rem;\">No upcoming events in the next 7 days.</p>"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		for _, event := range events {
			Render(w, "calendar-event-card.html", event)
		}
		return
	}

	// Return JSON for API calls
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// listEventRecords fetches all calendar events using direct XRPC call
func (h *CalendarHandler) ListEventRecords(ctx context.Context, sess *bskyoauth.Session) ([]*models.CalendarEvent, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		sess.PDS, sess.DID, CalendarEventCollection)

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
			Cid   string                 `json:"cid"`
			Value map[string]interface{} `json:"value"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: Received %d calendar event records from AT Protocol\n", len(result.Records))

	// Convert to CalendarEvent models
	events := make([]*models.CalendarEvent, 0, len(result.Records))
	for _, record := range result.Records {
		event, err := models.ParseCalendarEvent(record.Value, record.Uri, record.Cid)
		if err != nil {
			// Log error but continue with other events
			fmt.Printf("WARNING: Failed to parse calendar event %s: %v\n", record.Uri, err)
			continue
		}
		events = append(events, event)
	}

	fmt.Printf("DEBUG: Fetched %d calendar events from repository\n", len(events))

	return events, nil
}

// getEventRecord fetches a single event using direct XRPC call
func (h *CalendarHandler) getEventRecord(ctx context.Context, sess *bskyoauth.Session, rkey string) (*models.CalendarEvent, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, sess.DID, CalendarEventCollection, rkey)

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
		Uri   string                 `json:"uri"`
		Cid   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to CalendarEvent model
	event, err := models.ParseCalendarEvent(result.Value, result.Uri, result.Cid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	return event, nil
}

// listEventsFromRSVPs fetches events that the user has RSVP'd to
func (h *CalendarHandler) ListEventsFromRSVPs(ctx context.Context, sess *bskyoauth.Session) ([]*models.CalendarEvent, error) {
	// First, fetch all RSVPs
	rsvps, err := h.listRSVPRecords(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSVPs: %w", err)
	}

	fmt.Printf("DEBUG: Found %d RSVPs\n", len(rsvps))

	// Extract unique event URIs from RSVPs
	eventURIs := make(map[string]bool)
	for _, rsvp := range rsvps {
		if rsvp.Subject != nil && rsvp.Subject.URI != "" {
			eventURIs[rsvp.Subject.URI] = true
		}
	}

	fmt.Printf("DEBUG: Found %d unique event URIs from RSVPs\n", len(eventURIs))

	// Fetch each event by URI
	events := make([]*models.CalendarEvent, 0, len(eventURIs))
	for eventURI := range eventURIs {
		event, err := h.getEventByURI(ctx, sess, eventURI)
		if err != nil {
			fmt.Printf("WARNING: Failed to fetch event %s: %v\n", eventURI, err)
			continue
		}
		events = append(events, event)
	}

	fmt.Printf("DEBUG: Successfully fetched %d events from RSVPs\n", len(events))

	return events, nil
}

// getEventByURI fetches an event by its AT URI
func (h *CalendarHandler) getEventByURI(ctx context.Context, sess *bskyoauth.Session, uri string) (*models.CalendarEvent, error) {
	// Parse URI: at://did:plc:xxx/community.lexicon.calendar.event/rkey
	// Extract DID, collection, and rkey
	if len(uri) < 5 || uri[:5] != "at://" {
		return nil, fmt.Errorf("invalid AT URI: %s", uri)
	}

	parts := strings.Split(uri[5:], "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid AT URI format: %s", uri)
	}

	did := parts[0]
	collection := parts[1]
	rkey := parts[2]

	// Build XRPC URL - use the repo from the URI, not the session DID
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, did, collection, rkey)

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
		Uri   string                 `json:"uri"`
		Cid   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to CalendarEvent model
	event, err := models.ParseCalendarEvent(result.Value, result.Uri, result.Cid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	return event, nil
}

// listRSVPRecords fetches all RSVP records using direct XRPC call
func (h *CalendarHandler) listRSVPRecords(ctx context.Context, sess *bskyoauth.Session) ([]*models.CalendarRSVP, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		sess.PDS, sess.DID, CalendarRSVPCollection)

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
			Cid   string                 `json:"cid"`
			Value map[string]interface{} `json:"value"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to CalendarRSVP models
	rsvps := make([]*models.CalendarRSVP, 0, len(result.Records))
	for _, record := range result.Records {
		rsvp, err := models.ParseCalendarRSVP(record.Value, record.Uri, record.Cid)
		if err != nil {
			// Log error but continue with other RSVPs
			continue
		}
		rsvps = append(rsvps, rsvp)
	}

	return rsvps, nil
}

// sortEventsByDate sorts events by StartsAt in reverse chronological order (newest first)
// Events without StartsAt are placed at the end, sorted by CreatedAt
func sortEventsByDate(events []*models.CalendarEvent) {
	sort.Slice(events, func(i, j int) bool {
		eventI := events[i]
		eventJ := events[j]

		// If both have StartsAt, sort by StartsAt (newest first)
		if eventI.StartsAt != nil && eventJ.StartsAt != nil {
			return eventI.StartsAt.After(*eventJ.StartsAt)
		}

		// Events with StartsAt come before events without
		if eventI.StartsAt != nil {
			return true
		}
		if eventJ.StartsAt != nil {
			return false
		}

		// Both don't have StartsAt, sort by CreatedAt (newest first)
		return eventI.CreatedAt.After(eventJ.CreatedAt)
	})
}
