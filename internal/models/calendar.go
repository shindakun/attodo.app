package models

import (
	"fmt"
	"strings"
	"time"
)

// Event status constants
const (
	EventStatusPlanned     = "planned"     // Created but not finalized
	EventStatusScheduled   = "scheduled"   // Created and finalized
	EventStatusRescheduled = "rescheduled" // Event time/details changed
	EventStatusCancelled   = "cancelled"   // Event removed
	EventStatusPostponed   = "postponed"   // No new date set
)

// Attendance mode constants
const (
	AttendanceModeVirtual   = "virtual"    // Online only
	AttendanceModeInPerson  = "in-person"  // Physical location only
	AttendanceModeHybrid    = "hybrid"     // Both online and in-person
)

// RSVP status constants
const (
	RSVPStatusInterested = "interested" // Interested in the event
	RSVPStatusGoing      = "going"      // Going to the event
	RSVPStatusNotGoing   = "notgoing"   // Not going to the event
)

// CalendarEvent represents a community.lexicon.calendar.event record
type CalendarEvent struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartsAt    *time.Time `json:"startsAt,omitempty"`
	EndsAt      *time.Time `json:"endsAt,omitempty"`
	Mode        string     `json:"mode,omitempty"`        // hybrid, in-person, virtual
	Status      string     `json:"status,omitempty"`      // planned, scheduled, etc.
	Locations   []Location `json:"locations,omitempty"`
	URIs        []string   `json:"uris,omitempty"`

	// AT Protocol metadata
	RKey string `json:"rKey,omitempty"`
	URI  string `json:"uri,omitempty"`
	CID  string `json:"cid,omitempty"`
}

// Location represents a location where an event takes place
type Location struct {
	Name    string  `json:"name,omitempty"`
	Address string  `json:"address,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
}

// CalendarRSVP represents a community.lexicon.calendar.rsvp record
type CalendarRSVP struct {
	Subject *StrongRef `json:"subject"` // Reference to event
	Status  string     `json:"status"`  // interested, going, notgoing

	// AT Protocol metadata
	RKey string `json:"-"`
	URI  string `json:"-"`
	CID  string `json:"-"`
}

// StrongRef represents a com.atproto.repo.strongRef
type StrongRef struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// IsUpcoming returns true if the event starts in the future
func (e *CalendarEvent) IsUpcoming() bool {
	if e.StartsAt == nil {
		return false
	}
	return e.StartsAt.After(time.Now())
}

// IsPast returns true if the event has already ended
func (e *CalendarEvent) IsPast() bool {
	if e.EndsAt != nil {
		return e.EndsAt.Before(time.Now())
	}
	if e.StartsAt != nil {
		return e.StartsAt.Before(time.Now())
	}
	return false
}

// IsCancelled returns true if the event is cancelled or postponed
func (e *CalendarEvent) IsCancelled() bool {
	return e.Status == EventStatusCancelled || e.Status == EventStatusPostponed
}

// StartsWithin returns true if the event starts within the given duration
func (e *CalendarEvent) StartsWithin(d time.Duration) bool {
	if e.StartsAt == nil {
		return false
	}
	now := time.Now()
	return e.StartsAt.After(now) && e.StartsAt.Before(now.Add(d))
}

// FormatStatus returns a human-readable status string
func (e *CalendarEvent) FormatStatus() string {
	switch e.Status {
	case EventStatusPlanned:
		return "Planned"
	case EventStatusScheduled:
		return "Scheduled"
	case EventStatusRescheduled:
		return "Rescheduled"
	case EventStatusCancelled:
		return "Cancelled"
	case EventStatusPostponed:
		return "Postponed"
	case "": // Empty status
		return ""
	default:
		return ""
	}
}

// HasKnownStatus returns true if the event has a recognized status
func (e *CalendarEvent) HasKnownStatus() bool {
	switch e.Status {
	case EventStatusPlanned, EventStatusScheduled, EventStatusRescheduled, EventStatusCancelled, EventStatusPostponed:
		return true
	default:
		return false
	}
}

// FormatMode returns a human-readable attendance mode string
func (e *CalendarEvent) FormatMode() string {
	switch e.Mode {
	case AttendanceModeVirtual:
		return "Virtual"
	case AttendanceModeInPerson:
		return "In Person"
	case AttendanceModeHybrid:
		return "Hybrid"
	default:
		return ""
	}
}

// FormatRSVPStatus returns a human-readable RSVP status string
func (r *CalendarRSVP) FormatStatus() string {
	switch r.Status {
	case RSVPStatusInterested:
		return "Interested"
	case RSVPStatusGoing:
		return "Going"
	case RSVPStatusNotGoing:
		return "Not Going"
	default:
		return "Unknown"
	}
}

// ExtractDID extracts the DID from the event URI
// Example: at://did:plc:xxx/community.lexicon.calendar.event/abc123 -> did:plc:xxx
func (e *CalendarEvent) ExtractDID() string {
	if e.URI == "" {
		return ""
	}

	// Remove "at://" prefix
	uri := e.URI
	if len(uri) > 5 && uri[:5] == "at://" {
		uri = uri[5:]
	}

	// Find first slash to get DID
	slashIndex := -1
	for i := 0; i < len(uri); i++ {
		if uri[i] == '/' {
			slashIndex = i
			break
		}
	}

	if slashIndex > 0 {
		return uri[:slashIndex]
	}

	return ""
}

// SmokesignalURL returns the Smokesignal event URL if this is a Smokesignal event
// Returns empty string if not a Smokesignal event or if URI cannot be parsed
func (e *CalendarEvent) SmokesignalURL() string {
	did := e.ExtractDID()
	if did == "" || e.RKey == "" {
		return ""
	}

	// Smokesignal URL format: https://smokesignal.events/{did}/{rkey}
	return fmt.Sprintf("https://smokesignal.events/%s/%s", did, e.RKey)
}

// TruncatedDescription returns a truncated version of the description
func (e *CalendarEvent) TruncatedDescription(maxLen int) string {
	if len(e.Description) <= maxLen {
		return e.Description
	}
	return e.Description[:maxLen] + "..."
}

// ParseCalendarEvent parses an AT Protocol record into a CalendarEvent
func ParseCalendarEvent(record map[string]interface{}, uri, cid string) (*CalendarEvent, error) {
	event := &CalendarEvent{
		URI:  uri,
		CID:  cid,
		RKey: extractRKey(uri),
	}

	// Required: name
	if name, ok := record["name"].(string); ok {
		event.Name = name
	} else {
		return nil, fmt.Errorf("missing required field: name")
	}

	// Required: createdAt
	if createdAtStr, ok := record["createdAt"].(string); ok {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid createdAt: %w", err)
		}
		event.CreatedAt = createdAt
	} else {
		return nil, fmt.Errorf("missing required field: createdAt")
	}

	// Optional: description
	if description, ok := record["description"].(string); ok {
		event.Description = description
	}

	// Optional: startsAt
	if startsAtStr, ok := record["startsAt"].(string); ok {
		startsAt, err := time.Parse(time.RFC3339, startsAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid startsAt: %w", err)
		}
		event.StartsAt = &startsAt
	}

	// Optional: endsAt
	if endsAtStr, ok := record["endsAt"].(string); ok {
		endsAt, err := time.Parse(time.RFC3339, endsAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid endsAt: %w", err)
		}
		event.EndsAt = &endsAt
	}

	// Optional: mode
	if mode, ok := record["mode"].(string); ok {
		// Strip lexicon prefix if present (e.g., "community.lexicon.calendar.event#hybrid" -> "hybrid")
		if idx := strings.LastIndex(mode, "#"); idx != -1 {
			event.Mode = mode[idx+1:]
		} else {
			event.Mode = mode
		}
	}

	// Optional: status
	if status, ok := record["status"].(string); ok {
		// Strip lexicon prefix if present (e.g., "community.lexicon.calendar.event#scheduled" -> "scheduled")
		if idx := strings.LastIndex(status, "#"); idx != -1 {
			event.Status = status[idx+1:]
		} else {
			event.Status = status
		}
	}

	// Optional: locations
	if locationsRaw, ok := record["locations"].([]interface{}); ok {
		for _, locRaw := range locationsRaw {
			if locMap, ok := locRaw.(map[string]interface{}); ok {
				loc := Location{}
				if name, ok := locMap["name"].(string); ok {
					loc.Name = name
				}
				if address, ok := locMap["address"].(string); ok {
					loc.Address = address
				}
				if lat, ok := locMap["lat"].(float64); ok {
					loc.Lat = lat
				}
				if lon, ok := locMap["lon"].(float64); ok {
					loc.Lon = lon
				}
				event.Locations = append(event.Locations, loc)
			}
		}
	}

	// Optional: uris
	if urisRaw, ok := record["uris"].([]interface{}); ok {
		for _, uriRaw := range urisRaw {
			if uri, ok := uriRaw.(string); ok {
				event.URIs = append(event.URIs, uri)
			}
		}
	}

	return event, nil
}

// ParseCalendarRSVP parses an AT Protocol record into a CalendarRSVP
func ParseCalendarRSVP(record map[string]interface{}, uri, cid string) (*CalendarRSVP, error) {
	rsvp := &CalendarRSVP{
		URI:  uri,
		CID:  cid,
		RKey: extractRKey(uri),
	}

	// Required: subject
	if subjectRaw, ok := record["subject"].(map[string]interface{}); ok {
		subject := &StrongRef{}
		if subjectURI, ok := subjectRaw["uri"].(string); ok {
			subject.URI = subjectURI
		} else {
			return nil, fmt.Errorf("missing subject.uri")
		}
		if subjectCID, ok := subjectRaw["cid"].(string); ok {
			subject.CID = subjectCID
		} else {
			return nil, fmt.Errorf("missing subject.cid")
		}
		rsvp.Subject = subject
	} else {
		return nil, fmt.Errorf("missing required field: subject")
	}

	// Required: status
	if status, ok := record["status"].(string); ok {
		// Strip lexicon prefix if present (e.g., "community.lexicon.calendar.rsvp#going" -> "going")
		if idx := strings.LastIndex(status, "#"); idx != -1 {
			rsvp.Status = status[idx+1:]
		} else {
			rsvp.Status = status
		}
	} else {
		return nil, fmt.Errorf("missing required field: status")
	}

	return rsvp, nil
}

// extractRKey extracts the record key from an AT URI
// Example: at://did:plc:xxx/community.lexicon.calendar.event/abc123 -> abc123
func extractRKey(uri string) string {
	// Simple extraction - find last slash and return everything after it
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			return uri[i+1:]
		}
	}
	return ""
}
