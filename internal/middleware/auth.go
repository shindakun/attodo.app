package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/shindakun/attodo/internal/handlers"
	"github.com/shindakun/attodo/internal/session"
)

type AuthMiddleware struct {
	authHandler *handlers.AuthHandler
}

func NewAuthMiddleware(authHandler *handlers.AuthHandler) *AuthMiddleware {
	return &AuthMiddleware{authHandler: authHandler}
}

// RequireAuth ensures user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Middleware: Checking auth for %s", r.URL.Path)

		sess, err := m.authHandler.GetSession(r)
		if err != nil {
			log.Printf("Middleware: Auth failed, redirecting to /login: %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		log.Printf("Middleware: Auth successful for DID: %s", sess.DID)

		// Add session to context
		ctx := context.WithValue(r.Context(), session.SessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
