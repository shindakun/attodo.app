package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shindakun/attodo/internal/models"
)

// SupporterRepo handles supporter data persistence
type SupporterRepo struct {
	db *DB
}

// NewSupporterRepo creates a new supporter repository
func NewSupporterRepo(db *DB) *SupporterRepo {
	return &SupporterRepo{db: db}
}

// GetByDID retrieves a supporter by their DID
func (r *SupporterRepo) GetByDID(did string) (*models.Supporter, error) {
	var s models.Supporter
	err := r.db.QueryRow(`
		SELECT id, did, handle, email, stripe_customer_id, stripe_subscription_id,
		       plan_type, is_active, start_date, end_date, created_at, updated_at
		FROM supporters
		WHERE did = ?
	`, did).Scan(
		&s.ID, &s.DID, &s.Handle, &s.Email, &s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.PlanType, &s.IsActive, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not a supporter
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get supporter: %w", err)
	}

	return &s, nil
}

// GetByCustomerID retrieves a supporter by Stripe customer ID
func (r *SupporterRepo) GetByCustomerID(customerID string) (*models.Supporter, error) {
	var s models.Supporter
	err := r.db.QueryRow(`
		SELECT id, did, handle, email, stripe_customer_id, stripe_subscription_id,
		       plan_type, is_active, start_date, end_date, created_at, updated_at
		FROM supporters
		WHERE stripe_customer_id = ?
	`, customerID).Scan(
		&s.ID, &s.DID, &s.Handle, &s.Email, &s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.PlanType, &s.IsActive, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get supporter by customer ID: %w", err)
	}

	return &s, nil
}

// GetBySubscriptionID retrieves a supporter by Stripe subscription ID
func (r *SupporterRepo) GetBySubscriptionID(subscriptionID string) (*models.Supporter, error) {
	var s models.Supporter
	err := r.db.QueryRow(`
		SELECT id, did, handle, email, stripe_customer_id, stripe_subscription_id,
		       plan_type, is_active, start_date, end_date, created_at, updated_at
		FROM supporters
		WHERE stripe_subscription_id = ?
	`, subscriptionID).Scan(
		&s.ID, &s.DID, &s.Handle, &s.Email, &s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.PlanType, &s.IsActive, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get supporter by subscription ID: %w", err)
	}

	return &s, nil
}

// Create creates a new supporter record
func (r *SupporterRepo) Create(s *models.Supporter) error {
	result, err := r.db.Exec(`
		INSERT INTO supporters (did, handle, email, stripe_customer_id, stripe_subscription_id,
		                        plan_type, is_active, start_date, end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.DID, s.Handle, s.Email, s.StripeCustomerID, s.StripeSubscriptionID,
		s.PlanType, s.IsActive, s.StartDate, s.EndDate)

	if err != nil {
		return fmt.Errorf("failed to create supporter: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	s.ID = id
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	return nil
}

// Update updates an existing supporter record
func (r *SupporterRepo) Update(s *models.Supporter) error {
	s.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		UPDATE supporters
		SET handle = ?, email = ?, stripe_customer_id = ?, stripe_subscription_id = ?,
		    plan_type = ?, is_active = ?, start_date = ?, end_date = ?, updated_at = ?
		WHERE did = ?
	`, s.Handle, s.Email, s.StripeCustomerID, s.StripeSubscriptionID,
		s.PlanType, s.IsActive, s.StartDate, s.EndDate, s.UpdatedAt, s.DID)

	if err != nil {
		return fmt.Errorf("failed to update supporter: %w", err)
	}

	return nil
}

// Delete removes a supporter record
func (r *SupporterRepo) Delete(did string) error {
	_, err := r.db.Exec(`DELETE FROM supporters WHERE did = ?`, did)
	if err != nil {
		return fmt.Errorf("failed to delete supporter: %w", err)
	}
	return nil
}

// IsSupporter checks if a user is an active supporter
// Returns true if the user is active and within their subscription period
func (r *SupporterRepo) IsSupporter(did string) (bool, error) {
	var isActive bool
	err := r.db.QueryRow(`
		SELECT is_active
		FROM supporters
		WHERE did = ? AND (end_date IS NULL OR end_date > datetime('now'))
	`, did).Scan(&isActive)

	if err == sql.ErrNoRows {
		return false, nil // Not a supporter
	}
	if err != nil {
		return false, fmt.Errorf("failed to check supporter status: %w", err)
	}

	return isActive, nil
}

// CountActiveSupporter returns the number of active supporters
func (r *SupporterRepo) CountActiveSupporter() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM supporters
		WHERE is_active = 1 AND (end_date IS NULL OR end_date > datetime('now'))
	`).Scan(&count)

	return count, err
}
