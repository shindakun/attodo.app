package supporter

import (
	"fmt"
	"sync"
	"time"

	"github.com/shindakun/attodo/internal/database"
	"github.com/shindakun/attodo/internal/models"
)

// Service handles supporter business logic with caching
type Service struct {
	repo  *database.SupporterRepo
	cache map[string]*cachedStatus
	mu    sync.RWMutex
}

type cachedStatus struct {
	IsActive  bool
	ExpiresAt time.Time
}

// NewService creates a new supporter service
func NewService(repo *database.SupporterRepo) *Service {
	return &Service{
		repo:  repo,
		cache: make(map[string]*cachedStatus),
	}
}

// IsSupporter checks if a user is an active supporter (with caching)
func (s *Service) IsSupporter(did string) (bool, error) {
	// Check cache first
	s.mu.RLock()
	if cached, ok := s.cache[did]; ok && time.Now().Before(cached.ExpiresAt) {
		s.mu.RUnlock()
		return cached.IsActive, nil
	}
	s.mu.RUnlock()

	// Query database
	isActive, err := s.repo.IsSupporter(did)
	if err != nil {
		return false, err
	}

	// Cache result for 5 minutes
	s.mu.Lock()
	s.cache[did] = &cachedStatus{
		IsActive:  isActive,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	s.mu.Unlock()

	return isActive, nil
}

// ActivateSupporter creates or updates a supporter record
func (s *Service) ActivateSupporter(did, handle, email, customerID, subscriptionID string) error {
	existing, err := s.repo.GetByDID(did)
	if err != nil {
		return fmt.Errorf("failed to check existing supporter: %w", err)
	}

	if existing != nil {
		// Update existing record
		existing.Handle = handle
		existing.Email = email
		existing.StripeCustomerID = customerID
		existing.StripeSubscriptionID = subscriptionID
		existing.IsActive = true
		existing.EndDate = nil // Clear any end date

		if err := s.repo.Update(existing); err != nil {
			return fmt.Errorf("failed to update supporter: %w", err)
		}
	} else {
		// Create new record
		supporter := &models.Supporter{
			DID:                  did,
			Handle:               handle,
			Email:                email,
			StripeCustomerID:     customerID,
			StripeSubscriptionID: subscriptionID,
			PlanType:             "supporter",
			IsActive:             true,
			StartDate:            time.Now(),
		}

		if err := s.repo.Create(supporter); err != nil {
			return fmt.Errorf("failed to create supporter: %w", err)
		}
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, did)
	s.mu.Unlock()

	return nil
}

// DeactivateSupporter marks a supporter as inactive with grace period
func (s *Service) DeactivateSupporter(subscriptionID string, gracePeriodEnd time.Time) error {
	supporter, err := s.repo.GetBySubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get supporter: %w", err)
	}

	if supporter == nil {
		return fmt.Errorf("supporter not found for subscription: %s", subscriptionID)
	}

	supporter.EndDate = &gracePeriodEnd
	// Don't set IsActive to false yet - let it expire naturally

	if err := s.repo.Update(supporter); err != nil {
		return fmt.Errorf("failed to deactivate supporter: %w", err)
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, supporter.DID)
	s.mu.Unlock()

	return nil
}

// GetSupporter retrieves full supporter details
func (s *Service) GetSupporter(did string) (*models.Supporter, error) {
	return s.repo.GetByDID(did)
}

// GetByCustomerID retrieves supporter by Stripe customer ID
func (s *Service) GetByCustomerID(customerID string) (*models.Supporter, error) {
	return s.repo.GetByCustomerID(customerID)
}

// GetBySubscriptionID retrieves supporter by Stripe subscription ID
func (s *Service) GetBySubscriptionID(subscriptionID string) (*models.Supporter, error) {
	return s.repo.GetBySubscriptionID(subscriptionID)
}

// CountActiveSupporter returns the number of active supporters
func (s *Service) CountActiveSupporter() (int, error) {
	return s.repo.CountActiveSupporter()
}
