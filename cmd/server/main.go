package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shindakun/attodo/internal/config"
	"github.com/shindakun/attodo/internal/database"
	"github.com/shindakun/attodo/internal/handlers"
	"github.com/shindakun/attodo/internal/jobs"
	"github.com/shindakun/attodo/internal/middleware"
	"github.com/shindakun/attodo/internal/push"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database (runs migrations automatically)
	db, err := database.New(cfg.DBPath, cfg.MigrationsDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database initialized at %s", cfg.DBPath)

	// Initialize repositories
	notificationRepo := database.NewNotificationRepo(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	authMiddleware := middleware.NewAuthMiddleware(authHandler)
	taskHandler := handlers.NewTaskHandler(authHandler.Client())
	listHandler := handlers.NewListHandler(authHandler.Client())
	settingsHandler := handlers.NewSettingsHandler(authHandler.Client())
	pushHandler := handlers.NewPushHandler(notificationRepo)

	// Wire up cross-references between handlers
	taskHandler.SetListHandler(listHandler)

	// Initialize push notification sender (only if VAPID keys are configured)
	var pushSender *push.Sender
	var jobRunner *jobs.Runner
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" {
		pushSender = push.NewSender(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubscriber)
		pushHandler.SetSender(pushSender)
		log.Println("Push notification sender initialized")

		// Initialize background job runner (check every 5 minutes)
		jobRunner = jobs.NewRunner(5 * time.Minute)

		// Create and register notification check job
		notificationJob := jobs.NewNotificationCheckJob(notificationRepo, authHandler.Client(), pushSender)
		jobRunner.AddJob(notificationJob)

		// Start job runner
		jobRunner.Start()
		log.Println("Background job runner started")
	} else {
		log.Println("VAPID keys not configured - push notifications disabled")
		log.Println("Run 'go run ./cmd/vapid' to generate VAPID keys")
	}

	// Initialize templates
	handlers.InitTemplates(cfg)

	// Setup routes
	mux := http.NewServeMux()

	// Track registered routes for logging
	routes := []string{}
	logRoute := func(pattern string) {
		routes = append(routes, pattern)
	}

	// Service worker with custom scope header
	mux.HandleFunc("/static/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "static/sw.js")
	})
	logRoute("GET /static/sw.js")

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	logRoute("GET /static/*")

	// Health endpoint
	mux.HandleFunc("/health", handleHealth(cfg))
	logRoute("GET /health")

	// Public routes
	mux.HandleFunc("/", handleLanding(authHandler))
	logRoute("GET /")
	mux.HandleFunc("/oauth-client-metadata.json", authHandler.Client().ClientMetadataHandler())
	logRoute("GET /oauth-client-metadata.json")
	mux.HandleFunc("/login", authHandler.HandleLogin)
	logRoute("GET /login")
	mux.HandleFunc("/callback", authHandler.Client().CallbackHandler(authHandler.CallbackSuccess))
	logRoute("GET /callback")
	mux.HandleFunc("/logout", authHandler.Logout)
	logRoute("GET /logout")

	// Documentation routes
	mux.HandleFunc("/docs", handlers.Docs)
	logRoute("GET /docs")
	mux.HandleFunc("/docs/", handlers.DocsPage)
	logRoute("GET /docs/*")
	mux.HandleFunc("/docs/images/", handlers.DocsImage)
	logRoute("GET /docs/images/*")

	// Public list view route
	mux.HandleFunc("/list/", listHandler.HandlePublicListView)
	logRoute("GET /list/*")

	// Protected routes
	mux.Handle("/app", authMiddleware.RequireAuth(http.HandlerFunc(handleDashboard)))
	logRoute("GET /app [protected]")
	mux.Handle("/app/tasks", authMiddleware.RequireAuth(http.HandlerFunc(taskHandler.HandleTasks)))
	logRoute("GET/POST /app/tasks [protected]")
	mux.Handle("/app/lists", authMiddleware.RequireAuth(http.HandlerFunc(listHandler.HandleLists)))
	logRoute("GET/POST /app/lists [protected]")
	mux.Handle("/app/lists/view/", authMiddleware.RequireAuth(http.HandlerFunc(listHandler.HandleListDetail)))
	logRoute("GET /app/lists/view/* [protected]")
	mux.Handle("/app/settings", authMiddleware.RequireAuth(http.HandlerFunc(settingsHandler.HandleSettings)))
	logRoute("GET/PUT /app/settings [protected]")

	// Push notification routes
	mux.Handle("/app/push/vapid-key", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleGetVAPIDKey)))
	logRoute("GET /app/push/vapid-key [protected]")
	mux.Handle("/app/push/subscribe", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleSubscribe)))
	logRoute("POST /app/push/subscribe [protected]")
	mux.Handle("/app/push/unsubscribe", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleUnsubscribe)))
	logRoute("POST /app/push/unsubscribe [protected]")
	mux.Handle("/app/push/subscriptions", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleGetSubscriptions)))
	logRoute("GET /app/push/subscriptions [protected]")
	mux.Handle("/app/push/test", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleTestNotification)))
	logRoute("POST /app/push/test [protected]")
	mux.Handle("/app/push/check", authMiddleware.RequireAuth(http.HandlerFunc(pushHandler.HandleCheckTasks)))
	logRoute("POST /app/push/check [protected]")

	// Log all registered routes
	log.Println("Registered routes:")
	for _, route := range routes {
		log.Printf("  %s", route)
	}

	// Wrap with cache control middleware
	handler := middleware.NoCacheMiddleware(cfg, mux)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		if cfg.IsDev() {
			log.Printf("Starting server in DEVELOPMENT mode on :%s", cfg.Port)
			log.Printf("Cache headers disabled for development")
		} else {
			log.Printf("Starting server in PRODUCTION mode on :%s", cfg.Port)
		}
		log.Printf("Visit %s to get started", cfg.BaseURL)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down gracefully...")

	// Stop background jobs
	if jobRunner != nil {
		jobRunner.Stop()
	}

	log.Println("Shutdown complete")
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
