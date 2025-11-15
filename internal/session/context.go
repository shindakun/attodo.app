package session

import (
	"net/http"

	"github.com/shindakun/bskyoauth"
)

type contextKey string

const SessionKey contextKey = "session"

// GetSession extracts session from request context
func GetSession(r *http.Request) (*bskyoauth.Session, bool) {
	session, ok := r.Context().Value(SessionKey).(*bskyoauth.Session)
	return session, ok
}
