package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shindakun/attodo/internal/config"
	"github.com/shindakun/attodo/internal/handlers"
	"github.com/shindakun/attodo/internal/middleware"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	authMiddleware := middleware.NewAuthMiddleware(authHandler)
	taskHandler := handlers.NewTaskHandler(authHandler.Client())

	// Initialize templates
	handlers.InitTemplates()

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Health endpoint
	mux.HandleFunc("/health", handleHealth(cfg))

	// Public routes
	mux.HandleFunc("/", handleLanding(authHandler))
	mux.HandleFunc("/client-metadata.json", authHandler.Client().ClientMetadataHandler())
	mux.HandleFunc("/login", authHandler.HandleLogin)
	mux.HandleFunc("/callback", authHandler.Client().CallbackHandler(authHandler.CallbackSuccess))
	mux.HandleFunc("/logout", authHandler.Logout)

	// Protected routes
	mux.Handle("/app", authMiddleware.RequireAuth(http.HandlerFunc(handleDashboard)))
	mux.Handle("/app/tasks", authMiddleware.RequireAuth(http.HandlerFunc(taskHandler.HandleTasks)))

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	log.Printf("Visit %s to get started", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func handleLanding(authHandler *handlers.AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user has a session
		sessionCookie, err := r.Cookie("session_id")
		if err == nil {
			// Try to get session
			session, err := authHandler.Client().GetSession(sessionCookie.Value)
			if err == nil && session != nil {
				// User is logged in, redirect to dashboard
				http.Redirect(w, r, "/app", http.StatusSeeOther)
				return
			}
		}

		// Not logged in, show landing page
		handlers.Render(w, "landing.html", nil)
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	handlers.Render(w, "dashboard.html", nil)
}

func handleHealth(cfg *config.Config) http.HandlerFunc {
	startTime := time.Now()

	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
			"baseURL":   cfg.BaseURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}
}
