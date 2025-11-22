package database

import (
	"os"
	"testing"
	"time"

	"github.com/shindakun/attodo/internal/models"
)

func TestNotificationRepo(t *testing.T) {
	// Create temporary test database
	dbPath := "./test_notifications.db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	// Initialize database with migrations
	db, err := New(dbPath, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepo(db)

	// Test notification user operations
	t.Run("NotificationUser CRUD", func(t *testing.T) {
		testDID := "did:plc:test123"

		// Create user
		user := &models.NotificationUser{
			DID:                  testDID,
			NotificationsEnabled: true,
		}
		if err := repo.CreateNotificationUser(user); err != nil {
			t.Fatalf("Failed to create notification user: %v", err)
		}

		// Get user
		fetched, err := repo.GetNotificationUser(testDID)
		if err != nil {
			t.Fatalf("Failed to get notification user: %v", err)
		}
		if fetched == nil {
			t.Fatal("Expected user, got nil")
		}
		if fetched.DID != testDID {
			t.Errorf("Expected DID %s, got %s", testDID, fetched.DID)
		}
		if !fetched.NotificationsEnabled {
			t.Error("Expected notifications enabled")
		}

		// Update user
		fetched.NotificationsEnabled = false
		if err := repo.UpdateNotificationUser(fetched); err != nil {
			t.Fatalf("Failed to update notification user: %v", err)
		}

		// Verify update
		updated, err := repo.GetNotificationUser(testDID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updated.NotificationsEnabled {
			t.Error("Expected notifications disabled")
		}
	})

	// Test push subscription operations
	t.Run("PushSubscription CRUD", func(t *testing.T) {
		testDID := "did:plc:test456"
		testEndpoint := "https://push.example.com/test123"

		// Create notification user first
		user := &models.NotificationUser{
			DID:                  testDID,
			NotificationsEnabled: true,
		}
		if err := repo.CreateNotificationUser(user); err != nil {
			t.Fatalf("Failed to create notification user: %v", err)
		}

		// Create subscription
		sub := &models.PushSubscription{
			DID:        testDID,
			Endpoint:   testEndpoint,
			P256dhKey:  "test_p256dh_key",
			AuthSecret: "test_auth_secret",
			UserAgent:  "Test/1.0",
		}
		if err := repo.CreatePushSubscription(sub); err != nil {
			t.Fatalf("Failed to create push subscription: %v", err)
		}
		if sub.ID == 0 {
			t.Error("Expected ID to be set")
		}

		// Get subscription by endpoint
		fetched, err := repo.GetPushSubscription(testEndpoint)
		if err != nil {
			t.Fatalf("Failed to get push subscription: %v", err)
		}
		if fetched == nil {
			t.Fatal("Expected subscription, got nil")
		}
		if fetched.Endpoint != testEndpoint {
			t.Errorf("Expected endpoint %s, got %s", testEndpoint, fetched.Endpoint)
		}

		// Get subscriptions by DID
		subs, err := repo.GetPushSubscriptionsByDID(testDID)
		if err != nil {
			t.Fatalf("Failed to get subscriptions by DID: %v", err)
		}
		if len(subs) != 1 {
			t.Errorf("Expected 1 subscription, got %d", len(subs))
		}

		// Update last used
		time.Sleep(time.Millisecond * 10)
		if err := repo.UpdatePushSubscriptionLastUsed(testEndpoint); err != nil {
			t.Fatalf("Failed to update last used: %v", err)
		}

		// Delete subscription
		if err := repo.DeletePushSubscription(testEndpoint); err != nil {
			t.Fatalf("Failed to delete subscription: %v", err)
		}

		// Verify deletion
		deleted, err := repo.GetPushSubscription(testEndpoint)
		if err != nil {
			t.Fatalf("Error checking deleted subscription: %v", err)
		}
		if deleted != nil {
			t.Error("Expected subscription to be deleted")
		}
	})

	// Test notification history
	t.Run("NotificationHistory", func(t *testing.T) {
		testDID := "did:plc:test789"
		taskURI := "at://did:plc:test789/app.attodo.task/abc123"

		// Create notification user
		user := &models.NotificationUser{
			DID:                  testDID,
			NotificationsEnabled: true,
		}
		if err := repo.CreateNotificationUser(user); err != nil {
			t.Fatalf("Failed to create notification user: %v", err)
		}

		// Create history entry
		history := &models.NotificationHistory{
			DID:              testDID,
			TaskURI:          taskURI,
			NotificationType: "overdue",
			Status:           "sent",
		}
		if err := repo.CreateNotificationHistory(history); err != nil {
			t.Fatalf("Failed to create notification history: %v", err)
		}

		// Check recent notification (should find it)
		recent, err := repo.GetRecentNotification(testDID, taskURI, 12)
		if err != nil {
			t.Fatalf("Failed to get recent notification: %v", err)
		}
		if recent == nil {
			t.Error("Expected to find recent notification")
		}

		// Check with short cooldown (should not find it)
		recent2, err := repo.GetRecentNotification(testDID, taskURI, 0)
		if err != nil {
			t.Fatalf("Failed to check notification: %v", err)
		}
		if recent2 != nil {
			t.Error("Should not find notification with 0 hour cooldown")
		}
	})
}
