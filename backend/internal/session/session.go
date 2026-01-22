package session

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

const sessionKey ctxKey = "session_id"

const CookieName = "session_id"

func Middleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := getOrCreateSessionID(w, r)
			_ = upsertSession(db, sessionID)
			ctx := context.WithValue(r.Context(), sessionKey, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetSessionID(r *http.Request) string {
	if val, ok := r.Context().Value(sessionKey).(string); ok && val != "" {
		return val
	}
	return ""
}

func getOrCreateSessionID(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(CookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	sessionID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30,
	})
	return sessionID
}

func upsertSession(db *sql.DB, sessionID string) error {
	now := time.Now().UTC()
	_, err := db.Exec(`
		INSERT INTO user_sessions (session_id, created_at, last_activity)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE last_activity = VALUES(last_activity)
	`, sessionID, now, now)
	return err
}
