package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shindakun/attodo/internal/session"
	stripeClient "github.com/shindakun/attodo/internal/stripe"
	"github.com/shindakun/attodo/internal/supporter"
	"github.com/stripe/stripe-go/v84"
)

// SupporterHandler handles supporter-related HTTP requests
type SupporterHandler struct {
	service      *supporter.Service
	stripeClient *stripeClient.Client
	baseURL      string
}

// NewSupporterHandler creates a new supporter handler
func NewSupporterHandler(service *supporter.Service, stripe *stripeClient.Client, baseURL string) *SupporterHandler {
	return &SupporterHandler{
		service:      service,
		stripeClient: stripe,
		baseURL:      baseURL,
	}
}

// HandleGetStatus returns the current supporter status for the logged-in user
// GET /supporter/status
func (h *SupporterHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isSupporter, err := h.service.IsSupporter(sess.DID)
	if err != nil {
		log.Printf("Failed to check supporter status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"isSupporter": isSupporter,
	})
}

// HandleCreateCheckoutSession creates a Stripe Checkout session
// GET /supporter/checkout
func (h *SupporterHandler) HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Build success and cancel URLs
	successURL := fmt.Sprintf("%s/app?supporter=success", h.baseURL)
	cancelURL := fmt.Sprintf("%s/app?supporter=cancelled", h.baseURL)

	// Note: Handle and email will be extracted from profile or set in webhook
	// The DID is the primary identifier we need

	// Create Stripe checkout session
	checkoutSession, err := h.stripeClient.CreateCheckoutSession(
		sess.DID,
		"", // handle - will be updated from webhook metadata if available
		"", // email - will be updated from webhook metadata if available
		successURL,
		cancelURL,
	)
	if err != nil {
		log.Printf("Failed to create checkout session: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	// Return the checkout session URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url": checkoutSession.URL,
	})
}

// HandleCreatePortalSession creates a Stripe Customer Portal session
// GET /supporter/portal
func (h *SupporterHandler) HandleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get supporter record to find Stripe customer ID
	supporter, err := h.service.GetSupporter(sess.DID)
	if err != nil {
		log.Printf("Failed to get supporter: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if supporter == nil || supporter.StripeCustomerID == "" {
		http.Error(w, "No subscription found", http.StatusNotFound)
		return
	}

	// Create portal session
	returnURL := fmt.Sprintf("%s/app", h.baseURL)
	portalSession, err := h.stripeClient.CreateCustomerPortalSession(supporter.StripeCustomerID, returnURL)
	if err != nil {
		log.Printf("Failed to create portal session: %v", err)
		http.Error(w, "Failed to create portal session", http.StatusInternalServerError)
		return
	}

	// Return the portal session URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url": portalSession.URL,
	})
}

// HandleStripeWebhook handles Stripe webhook events
// POST /supporter/webhook
func (h *SupporterHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	event, err := h.stripeClient.ConstructWebhookEvent(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "Webhook signature verification failed", http.StatusBadRequest)
		return
	}

	log.Printf("[Stripe Webhook] Received event: %s", event.Type)

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		h.handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		h.handlePaymentFailed(event)
	default:
		log.Printf("[Stripe Webhook] Unhandled event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutCompleted processes checkout.session.completed event
func (h *SupporterHandler) handleCheckoutCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("[Stripe Webhook] Error parsing checkout session: %v", err)
		return
	}

	// Get user info from session metadata
	did := session.Metadata["did"]
	handle := session.Metadata["handle"]
	email := session.Metadata["email"]

	if did == "" {
		log.Printf("[Stripe Webhook] Missing DID in checkout session metadata")
		return
	}

	// Extract customer and subscription IDs
	customerID := ""
	if session.Customer != nil {
		customerID = session.Customer.ID
	}

	subscriptionID := ""
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID
	}

	// Activate supporter status
	err := h.service.ActivateSupporter(did, handle, email, customerID, subscriptionID)
	if err != nil {
		log.Printf("[Stripe Webhook] Error activating supporter: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Activated supporter: %s (%s)", handle, did)
}

// handleSubscriptionCreated processes customer.subscription.created event
func (h *SupporterHandler) handleSubscriptionCreated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("[Stripe Webhook] Error parsing subscription: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Subscription created: %s for customer %s", subscription.ID, subscription.Customer.ID)
	// Usually handled by checkout.session.completed, but log for tracking
}

// handleSubscriptionUpdated processes customer.subscription.updated event
func (h *SupporterHandler) handleSubscriptionUpdated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("[Stripe Webhook] Error parsing subscription: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Subscription updated: %s (status: %s)", subscription.ID, subscription.Status)

	// If subscription is canceled or past_due, we might want to update supporter status
	if subscription.Status == stripe.SubscriptionStatusCanceled ||
		subscription.Status == stripe.SubscriptionStatusUnpaid {
		// Handle subscription end with grace period
		// In Stripe API v84+, period end is on subscription items, not the subscription
		var endDate time.Time
		if len(subscription.Items.Data) > 0 {
			endDate = time.Unix(subscription.Items.Data[0].CurrentPeriodEnd, 0)
		} else {
			// Fallback to now if no items found
			endDate = time.Now()
		}
		if err := h.service.DeactivateSupporter(subscription.ID, endDate); err != nil {
			log.Printf("[Stripe Webhook] Error deactivating supporter: %v", err)
		}
	}
}

// handleSubscriptionDeleted processes customer.subscription.deleted event
func (h *SupporterHandler) handleSubscriptionDeleted(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("[Stripe Webhook] Error parsing subscription: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Subscription deleted: %s", subscription.ID)

	// Deactivate supporter immediately (already past grace period)
	if err := h.service.DeactivateSupporter(subscription.ID, time.Now()); err != nil {
		log.Printf("[Stripe Webhook] Error deactivating supporter: %v", err)
	}
}

// handlePaymentSucceeded processes invoice.payment_succeeded event
func (h *SupporterHandler) handlePaymentSucceeded(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[Stripe Webhook] Error parsing invoice: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Payment succeeded: %s for customer %s (amount: %d %s)",
		invoice.ID, invoice.Customer.ID, invoice.AmountPaid, invoice.Currency)
}

// handlePaymentFailed processes invoice.payment_failed event
func (h *SupporterHandler) handlePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[Stripe Webhook] Error parsing invoice: %v", err)
		return
	}

	log.Printf("[Stripe Webhook] Payment failed: %s for customer %s", invoice.ID, invoice.Customer.ID)
	// Stripe will retry payments automatically, so we just log this for monitoring
}
