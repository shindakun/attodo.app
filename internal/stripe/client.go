package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v84"
	billingSession "github.com/stripe/stripe-go/v84/billingportal/session"
	checkoutSession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

// Client wraps Stripe API client with our configuration
type Client struct {
	webhookSecret string
	priceID       string
}

// NewClient creates a new Stripe client
// secretKey: Stripe secret key (sk_test_... or sk_live_...)
// webhookSecret: Stripe webhook signing secret (whsec_...)
// priceID: Stripe price ID for the supporter plan (price_...)
func NewClient(secretKey, webhookSecret, priceID string) *Client {
	stripe.Key = secretKey
	return &Client{
		webhookSecret: webhookSecret,
		priceID:       priceID,
	}
}

// CreateCheckoutSession creates a new Stripe Checkout session for subscription
func (c *Client) CreateCheckoutSession(did, handle, email, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(c.priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		BrandingSettings: &stripe.CheckoutSessionBrandingSettingsParams{
			BackgroundColor: stripe.String("#1a1a1a"),
			ButtonColor:     stripe.String("#5469d4"),
			BorderStyle:     stripe.String("rounded"),
		},
	}

	// Add customer email if provided
	if email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	// Add metadata for webhook processing
	params.AddMetadata("did", did)
	params.AddMetadata("handle", handle)
	if email != "" {
		params.AddMetadata("email", email)
	}

	sess, err := checkoutSession.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess, nil
}

// CreateCustomerPortalSession creates a Stripe Customer Portal session
// This allows users to manage their subscription (cancel, update payment, etc.)
func (c *Client) CreateCustomerPortalSession(customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := billingSession.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create portal session: %w", err)
	}

	return sess, nil
}

// ConstructWebhookEvent verifies and constructs a webhook event from the request
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	return event, nil
}
