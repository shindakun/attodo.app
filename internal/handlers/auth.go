package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/shindakun/attodo/internal/config"
	"github.com/shindakun/bskyoauth"
)

type AuthHandler struct {
	client *bskyoauth.Client
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	// Initialize bskyoauth with explicit scopes
	client := bskyoauth.NewClientWithOptions(bskyoauth.ClientOptions{
		BaseURL:    cfg.BaseURL,
		ClientName: cfg.ClientName,
		Scopes:     []string{"atproto", "transition:generic"},
	})

	return &AuthHandler{
		client: client,
	}
}

// Client returns the bskyoauth client for registering handlers
func (h *AuthHandler) Client() *bskyoauth.Client {
	return h.client
}

// HandleLogin wraps LoginHandler to show form when no handle provided
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		// Show login form (render the landing page)
		Render(w, "landing.html", nil)
		return
	}

	// Delegate to bskyoauth's LoginHandler
	h.client.LoginHandler()(w, r)
}

// CallbackSuccess is called after successful OAuth
func (h *AuthHandler) CallbackSuccess(w http.ResponseWriter, r *http.Request, sessionID string) {
	log.Printf("OAuth callback success, sessionID: %s", sessionID)

	// Store sessionID in simple cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30, // 30 days
	})

	// Redirect to home, not /app (avoid redirect loop)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout deletes the session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Delete from bskyoauth's session store
		h.client.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GetSession retrieves and refreshes the OAuth session
func (h *AuthHandler) GetSession(r *http.Request) (*bskyoauth.Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}

	session, err := h.client.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	// Refresh if needed
	if session.IsAccessTokenExpired(5 * time.Minute) {
		session, err = h.client.RefreshToken(r.Context(), session)
		if err != nil {
			return nil, err
		}
		h.client.UpdateSession(cookie.Value, session)
	}

	return session, nil
}
