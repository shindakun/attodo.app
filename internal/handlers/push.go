package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shindakun/attodo/internal/database"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/push"
	"github.com/shindakun/attodo/internal/session"
)

// PushHandler handles push notification subscription endpoints
type PushHandler struct {
	repo   *database.NotificationRepo
	sender *push.Sender
}

// NewPushHandler creates a new push notification handler
func NewPushHandler(repo *database.NotificationRepo) *PushHandler {
	return &PushHandler{
		repo:   repo,
		sender: nil, // Set later with SetSender()
	}
}

// SetSender sets the push notification sender (called after VAPID keys are loaded)
func (h *PushHandler) SetSender(sender *push.Sender) {
	h.sender = sender
}

// HandleGetVAPIDKey returns the public VAPID key for push subscription
func (h *PushHandler) HandleGetVAPIDKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.sender == nil {
		http.Error(w, "Push notifications not configured", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"publicKey": h.sender.PublicKey(),
	})
}

// SubscribeRequest represents a push subscription request from the client
type SubscribeRequest struct {
	Endpoint string            `json:"endpoint"`
	Keys     map[string]string `json:"keys"` // Contains p256dh and auth
}

// HandleSubscribe handles POST /app/push/subscribe
// Subscribes the current user to push notifications
func (h *PushHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user session
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse subscription request
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode subscribe request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Endpoint == "" {
		http.Error(w, "Missing endpoint", http.StatusBadRequest)
		return
	}
	p256dh, ok := req.Keys["p256dh"]
	if !ok || p256dh == "" {
		http.Error(w, "Missing p256dh key", http.StatusBadRequest)
		return
	}
	auth, ok := req.Keys["auth"]
	if !ok || auth == "" {
		http.Error(w, "Missing auth key", http.StatusBadRequest)
		return
	}

	// Ensure notification user exists BEFORE creating subscription (foreign key constraint)
	user, err := h.repo.GetNotificationUser(sess.DID)
	if err != nil {
		log.Printf("Failed to get notification user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Create notification user
		user = &models.NotificationUser{
			DID:                  sess.DID,
			NotificationsEnabled: true,
		}
		if err := h.repo.CreateNotificationUser(user); err != nil {
			log.Printf("Failed to create notification user: %v", err)
			http.Error(w, "Failed to enable notifications", http.StatusInternalServerError)
			return
		}
	} else if !user.NotificationsEnabled {
		// Enable notifications
		user.NotificationsEnabled = true
		if err := h.repo.UpdateNotificationUser(user); err != nil {
			log.Printf("Failed to enable notifications: %v", err)
		}
	}

	// Check if subscription already exists
	existing, err := h.repo.GetPushSubscription(req.Endpoint)
	if err != nil {
		log.Printf("Failed to check existing subscription: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existing != nil {
		// Update last used timestamp
		if err := h.repo.UpdatePushSubscriptionLastUsed(req.Endpoint); err != nil {
			log.Printf("Failed to update subscription last used: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Subscription already exists",
		})
		return
	}

	// Create new push subscription
	sub := &models.PushSubscription{
		DID:        sess.DID,
		Endpoint:   req.Endpoint,
		P256dhKey:  p256dh,
		AuthSecret: auth,
		UserAgent:  r.Header.Get("User-Agent"),
	}

	if err := h.repo.CreatePushSubscription(sub); err != nil {
		log.Printf("Failed to create push subscription for DID %s: %v", sess.DID, err)
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("Push subscription created for DID: %s", sess.DID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Subscription created successfully",
	})
}

// HandleUnsubscribe handles POST /app/push/unsubscribe
// Unsubscribes the current user from push notifications
func (h *PushHandler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user session
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse unsubscribe request
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode unsubscribe request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" {
		http.Error(w, "Missing endpoint", http.StatusBadRequest)
		return
	}

	// Verify the subscription belongs to this user
	existing, err := h.repo.GetPushSubscription(req.Endpoint)
	if err != nil {
		log.Printf("Failed to get subscription: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existing == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Subscription not found",
		})
		return
	}

	if existing.DID != sess.DID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete the subscription
	if err := h.repo.DeletePushSubscription(req.Endpoint); err != nil {
		log.Printf("Failed to delete subscription: %v", err)
		http.Error(w, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	// Check if user has any remaining subscriptions
	subs, err := h.repo.GetPushSubscriptionsByDID(sess.DID)
	if err != nil {
		log.Printf("Failed to check remaining subscriptions: %v", err)
	} else if len(subs) == 0 {
		// No more subscriptions, disable notifications
		user, err := h.repo.GetNotificationUser(sess.DID)
		if err != nil {
			log.Printf("Failed to get notification user: %v", err)
		} else if user != nil {
			user.NotificationsEnabled = false
			if err := h.repo.UpdateNotificationUser(user); err != nil {
				log.Printf("Failed to disable notifications: %v", err)
			}
		}
	}

	log.Printf("Push subscription deleted for DID: %s", sess.DID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Subscription deleted successfully",
	})
}

// HandleGetSubscriptions handles GET /app/push/subscriptions
// Returns all push subscriptions for the current user
func (h *PushHandler) HandleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user session
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all subscriptions for this user
	subs, err := h.repo.GetPushSubscriptionsByDID(sess.DID)
	if err != nil {
		log.Printf("Failed to get subscriptions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get notification user settings
	user, err := h.repo.GetNotificationUser(sess.DID)
	if err != nil {
		log.Printf("Failed to get notification user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	enabled := user != nil && user.NotificationsEnabled

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":       enabled,
		"subscriptions": subs,
	})
}

// HandleTestNotification handles POST /app/push/test
// Sends a test push notification to the current user
func (h *PushHandler) HandleTestNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user session
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if push sender is configured
	if h.sender == nil {
		http.Error(w, "Push notifications not configured", http.StatusServiceUnavailable)
		return
	}

	// Get user's push subscriptions
	subs, err := h.repo.GetPushSubscriptionsByDID(sess.DID)
	if err != nil {
		log.Printf("Failed to get subscriptions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(subs) == 0 {
		http.Error(w, "No push subscriptions found", http.StatusNotFound)
		return
	}

	// Create test notification
	notification := &push.Notification{
		Title: "AT Todo Test",
		Body:  "Push notifications are working! 🎉",
		Icon:  "/static/icon-192.png",
		Badge: "/static/icon-192.png",
		Tag:   "test",
		Data: map[string]interface{}{
			"type": "test",
		},
	}

	// Send to all subscriptions
	successCount, errors := h.sender.SendToAll(subs, notification)

	log.Printf("Test notification sent to %d/%d subscriptions for DID: %s", successCount, len(subs), sess.DID)

	if len(errors) > 0 {
		log.Printf("Test notification errors: %v", errors)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      successCount > 0,
		"sent":         successCount,
		"total":        len(subs),
		"errors":       len(errors),
		"errorDetails": errors,
	})
}

// HandleCheckTasks handles POST /app/push/check
// Checks user's tasks and sends notifications for due tasks
func (h *PushHandler) HandleCheckTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user session
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		// Service worker periodic sync may not have session
		// Just return OK silently
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check if push sender is configured
	if h.sender == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Get user's push subscriptions
	subs, err := h.repo.GetPushSubscriptionsByDID(sess.DID)
	if err != nil {
		log.Printf("Failed to get subscriptions: %v", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if len(subs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Log that periodic check was triggered
	log.Printf("[Push] Periodic check triggered for %s", sess.DID)

	// For now, the service worker handles task checking client-side
	// This endpoint exists for future server-side task checking
	w.WriteHeader(http.StatusNoContent)
}
