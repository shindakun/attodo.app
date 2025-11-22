package models

import "time"

// Supporter represents a user who has subscribed to support AT Todo
type Supporter struct {
	ID                   int64      `db:"id" json:"id"`
	DID                  string     `db:"did" json:"did"`
	Handle               string     `db:"handle" json:"handle"`
	Email                string     `db:"email" json:"email,omitempty"`
	StripeCustomerID     string     `db:"stripe_customer_id" json:"stripeCustomerId,omitempty"`
	StripeSubscriptionID string     `db:"stripe_subscription_id" json:"stripeSubscriptionId,omitempty"`
	PlanType             string     `db:"plan_type" json:"planType"`
	IsActive             bool       `db:"is_active" json:"isActive"`
	StartDate            time.Time  `db:"start_date" json:"startDate"`
	EndDate              *time.Time `db:"end_date" json:"endDate,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
}
