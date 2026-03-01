package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserContextKey contextKey = "user"

type User struct {
	ID             int
	OAuthProvider  string
	OAuthID        string
	Email          string
	Name           string
	AvatarURL      string
	TelegramChatID string
	NotifyTelegram bool
	NotifyEmail    bool
	IsAdmin        bool
	IsActive       bool
	CreatedAt      string
	LastSeenAt     string
}

// GetUser extracts the authenticated user from request context.
func GetUser(r *http.Request) *User {
	u, _ := r.Context().Value(UserContextKey).(*User)
	return u
}

// RequireAuth middleware: reads "session" cookie, validates the signed token,
// looks up session in DB, verifies not expired, loads user, sets user in context.
// Implements sliding session: if session expires within the next 24 hours, extend by 30 days.
// Updates sessions.last_seen_at and users.last_seen_at on each request.
// Returns 401 JSON error if not authenticated.
func RequireAuth(db *sql.DB, sessionSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}

			parts := strings.SplitN(cookie.Value, ":", 2)
			if len(parts) != 2 {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid session"})
				return
			}
			token, sig := parts[0], parts[1]

			expectedSig := computeHMAC(token, sessionSecret)
			if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid session"})
				return
			}

			var userID int
			var expiresAtStr string
			err = db.QueryRow(
				"SELECT user_id, expires_at FROM sessions WHERE token = ?", token,
			).Scan(&userID, &expiresAtStr)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session not found"})
				return
			}

			expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid session"})
				return
			}

			now := time.Now().UTC()
			if now.After(expiresAt) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session expired"})
				return
			}

			// Sliding session: extend if within last 24 hours before expiry
			if expiresAt.Sub(now) < 24*time.Hour {
				newExpiry := now.Add(30 * 24 * time.Hour)
				_, _ = db.Exec(
					"UPDATE sessions SET expires_at = ?, last_seen_at = ? WHERE token = ?",
					newExpiry.Format("2006-01-02 15:04:05"),
					now.Format("2006-01-02 15:04:05"),
					token,
				)
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    token + ":" + sig,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					Expires:  newExpiry,
				})
			} else {
				_, _ = db.Exec(
					"UPDATE sessions SET last_seen_at = ? WHERE token = ?",
					now.Format("2006-01-02 15:04:05"),
					token,
				)
			}

			// Update users.last_seen_at
			_, _ = db.Exec(
				"UPDATE users SET last_seen_at = ? WHERE id = ?",
				now.Format("2006-01-02 15:04:05"),
				userID,
			)

			var user User
			var email, name, avatarURL, telegramChatID, lastSeenAt sql.NullString
			err = db.QueryRow(
				`SELECT id, oauth_provider, oauth_id, email, name, avatar_url,
				        telegram_chat_id, notify_telegram, notify_email,
				        is_admin, is_active, created_at, last_seen_at
				 FROM users WHERE id = ?`, userID,
			).Scan(
				&user.ID, &user.OAuthProvider, &user.OAuthID,
				&email, &name, &avatarURL,
				&telegramChatID, &user.NotifyTelegram, &user.NotifyEmail,
				&user.IsAdmin, &user.IsActive, &user.CreatedAt, &lastSeenAt,
			)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
				return
			}
			user.Email = email.String
			user.Name = name.String
			user.AvatarURL = avatarURL.String
			user.TelegramChatID = telegramChatID.String
			user.LastSeenAt = lastSeenAt.String

			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware: checks that the user from context has is_admin=true.
// Returns 403 JSON error if not admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil || !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CreateSession creates a new session in the DB, returns the signed cookie value.
func CreateSession(db *sql.DB, userID int, sessionSecret string) (cookieValue string, expiresAt time.Time, err error) {
	token, err := generateUUID()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generating session token: %w", err)
	}

	expiresAt = time.Now().UTC().Add(30 * 24 * time.Hour)
	_, err = db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("inserting session: %w", err)
	}

	sig := computeHMAC(token, sessionSecret)
	return token + ":" + sig, expiresAt, nil
}

func computeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Set version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
