package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/session"
	"github.com/shindakun/bskyoauth"
)

const SettingsCollection = "app.attodo.settings"
const SettingsRKey = "settings" // Single record per user

type SettingsHandler struct {
	client *bskyoauth.Client
}

func NewSettingsHandler(client *bskyoauth.Client) *SettingsHandler {
	return &SettingsHandler{client: client}
}

// HandleSettings handles settings CRUD operations
func (h *SettingsHandler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetSettings(w, r)
	case http.MethodPut:
		h.handleUpdateSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSettings retrieves user settings
func (h *SettingsHandler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to get existing settings
	var record map[string]interface{}
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		var fetchErr error
		record, fetchErr = h.getRecord(r.Context(), s, SettingsRKey)
		return fetchErr
	})

	var settings *models.NotificationSettings

	if err != nil {
		// No settings record exists yet, return defaults
		log.Printf("No settings found, returning defaults: %v", err)
		settings = models.DefaultNotificationSettings()
	} else {
		// Parse existing settings
		settings = parseSettingsRecord(record)
		settings.RKey = SettingsRKey
		settings.URI = fmt.Sprintf("at://%s/%s/%s", sess.DID, SettingsCollection, SettingsRKey)
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Return settings as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// handleUpdateSettings updates user settings
func (h *SettingsHandler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var settings models.NotificationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set metadata
	settings.UpdatedAt = time.Now().UTC()

	// Convert to map for AT Protocol
	record := map[string]interface{}{
		"$type":             SettingsCollection,
		"notifyOverdue":     settings.NotifyOverdue,
		"notifyToday":       settings.NotifyToday,
		"notifySoon":        settings.NotifySoon,
		"hoursBefore":       settings.HoursBefore,
		"checkFrequency":    settings.CheckFrequency,
		"quietHoursEnabled": settings.QuietHoursEnabled,
		"quietStart":        settings.QuietStart,
		"quietEnd":          settings.QuietEnd,
		"pushEnabled":       settings.PushEnabled,
		"updatedAt":         settings.UpdatedAt.Format(time.RFC3339),
	}

	// Include appUsageHours if present
	if settings.AppUsageHours != nil {
		record["appUsageHours"] = settings.AppUsageHours
	}

	// Check if settings record already exists
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		_, fetchErr := h.getRecord(r.Context(), s, SettingsRKey)
		return fetchErr
	})

	if err != nil {
		// Create new settings record
		log.Printf("Creating new settings record")
		sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
			return h.createRecord(r.Context(), s, SettingsRKey, record)
		})
	} else {
		// Update existing settings record
		log.Printf("Updating existing settings record")
		sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
			return h.updateRecord(r.Context(), s, SettingsRKey, record)
		})
	}

	if err != nil {
		log.Printf("Failed to save settings: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save settings: %v", err), http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Return updated settings
	settings.RKey = SettingsRKey
	settings.URI = fmt.Sprintf("at://%s/%s/%s", sess.DID, SettingsCollection, SettingsRKey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// parseSettingsRecord parses a settings record from AT Protocol
func parseSettingsRecord(record map[string]interface{}) *models.NotificationSettings {
	settings := models.DefaultNotificationSettings()

	if v, ok := record["notifyOverdue"].(bool); ok {
		settings.NotifyOverdue = v
	}
	if v, ok := record["notifyToday"].(bool); ok {
		settings.NotifyToday = v
	}
	if v, ok := record["notifySoon"].(bool); ok {
		settings.NotifySoon = v
	}
	if v, ok := record["hoursBefore"].(float64); ok {
		settings.HoursBefore = int(v)
	}
	if v, ok := record["checkFrequency"].(float64); ok {
		settings.CheckFrequency = int(v)
	}
	if v, ok := record["quietHoursEnabled"].(bool); ok {
		settings.QuietHoursEnabled = v
	}
	if v, ok := record["quietStart"].(float64); ok {
		settings.QuietStart = int(v)
	}
	if v, ok := record["quietEnd"].(float64); ok {
		settings.QuietEnd = int(v)
	}
	if v, ok := record["pushEnabled"].(bool); ok {
		settings.PushEnabled = v
	}

	// Parse appUsageHours if present
	if usageMap, ok := record["appUsageHours"].(map[string]interface{}); ok {
		settings.AppUsageHours = make(map[string]int)
		for k, v := range usageMap {
			if count, ok := v.(float64); ok {
				settings.AppUsageHours[k] = int(count)
			}
		}
	}

	// Parse updatedAt
	if v, ok := record["updatedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			settings.UpdatedAt = t
		}
	}

	return settings
}

// getRecord retrieves a settings record using com.atproto.repo.getRecord
func (h *SettingsHandler) getRecord(ctx context.Context, sess *bskyoauth.Session, rkey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, sess.DID, SettingsCollection, rkey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	transport := bskyoauth.NewDPoPTransport(http.DefaultTransport, sess.DPoPKey, sess.AccessToken, sess.DPoPNonce)
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if dpopTransport, ok := transport.(bskyoauth.DPoPTransport); ok {
		sess.DPoPNonce = dpopTransport.GetNonce()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Value, nil
}

// createRecord creates a new settings record using com.atproto.repo.putRecord
func (h *SettingsHandler) createRecord(ctx context.Context, sess *bskyoauth.Session, rkey string, record map[string]interface{}) error {
	log.Printf("createRecord: DID=%s, Collection=%s, RKey=%s", sess.DID, SettingsCollection, rkey)

	if _, exists := record["$type"]; !exists {
		record["$type"] = SettingsCollection
	}

	body := map[string]interface{}{
		"repo":       sess.DID,
		"collection": SettingsCollection,
		"rkey":       rkey,
		"record":     record,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.putRecord", sess.PDS)
	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(strings.NewReader(string(bodyJSON))))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	transport := bskyoauth.NewDPoPTransport(http.DefaultTransport, sess.DPoPKey, sess.AccessToken, sess.DPoPNonce)
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if dpopTransport, ok := transport.(bskyoauth.DPoPTransport); ok {
		sess.DPoPNonce = dpopTransport.GetNonce()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// updateRecord updates an existing settings record using com.atproto.repo.putRecord
func (h *SettingsHandler) updateRecord(ctx context.Context, sess *bskyoauth.Session, rkey string, record map[string]interface{}) error {
	// For settings, update is the same as create since we use putRecord
	return h.createRecord(ctx, sess, rkey, record)
}

// WithRetry wraps an operation with session refresh on auth failure
func (h *SettingsHandler) WithRetry(ctx context.Context, sess *bskyoauth.Session, fn func(*bskyoauth.Session) error) (*bskyoauth.Session, error) {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		err := fn(sess)
		if err == nil {
			return sess, nil
		}

		// Check if it's a token expiration error
		if strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") {
			log.Printf("Token may be expired, attempting refresh (attempt %d/%d)", i+1, maxRetries)

			// Try to refresh the token
			newSess, refreshErr := h.client.RefreshToken(ctx, sess)
			if refreshErr != nil {
				log.Printf("Failed to refresh token: %v", refreshErr)
				return sess, err // Return original error
			}

			sess = newSess
			continue
		}

		// Not a token error, return immediately
		return sess, err
	}

	return sess, fmt.Errorf("max retries exceeded")
}
