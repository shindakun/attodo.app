package push

import (
	"encoding/json"
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/shindakun/attodo/internal/models"
)

// Sender handles sending Web Push notifications
type Sender struct {
	publicKey  string
	privateKey string
	subscriber string
}

// NewSender creates a new push notification sender
func NewSender(publicKey, privateKey, subscriber string) *Sender {
	return &Sender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subscriber: subscriber,
	}
}

// PublicKey returns the VAPID public key
func (s *Sender) PublicKey() string {
	return s.publicKey
}

// Notification represents a push notification payload
type Notification struct {
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Icon  string                 `json:"icon,omitempty"`
	Badge string                 `json:"badge,omitempty"`
	Tag   string                 `json:"tag,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// Send sends a push notification to a subscription
func (s *Sender) Send(sub *models.PushSubscription, notification *Notification) error {
	// Marshal notification to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create webpush subscription
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthSecret,
		},
	}

	// Send the notification
	resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             86400, // 24 hours
	})
	if err != nil {
		log.Printf("Push notification error for %s: %v", sub.Endpoint, err)
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != 201 {
		log.Printf("Push service returned status %d for %s", resp.StatusCode, sub.Endpoint)
		return fmt.Errorf("push service returned status %d", resp.StatusCode)
	}

	log.Printf("Push notification sent successfully (HTTP %d) to %s", resp.StatusCode, sub.Endpoint)
	return nil
}

// SendToAll sends a notification to all subscriptions for a DID
// Returns the number of successful sends and any errors encountered
func (s *Sender) SendToAll(subscriptions []*models.PushSubscription, notification *Notification) (int, []error) {
	var errors []error
	successCount := 0

	for _, sub := range subscriptions {
		if err := s.Send(sub, notification); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", sub.Endpoint, err))
		} else {
			successCount++
		}
	}

	return successCount, errors
}
