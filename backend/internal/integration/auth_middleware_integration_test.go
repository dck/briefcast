package integration_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dck/briefcast/internal/middleware"
	_ "modernc.org/sqlite"
)

func TestRequireAuthIntegration(t *testing.T) {
	t.Parallel()

	const (
		secret = "test-secret"
		token  = "session-token"
	)

	tests := []struct {
		name       string
		cookie     *http.Cookie
		wantStatus int
		wantUserID int
	}{
		{
			name: "success",
			cookie: &http.Cookie{
				Name:  "session",
				Value: token + ":" + signSession(token, secret),
			},
			wantStatus: http.StatusNoContent,
			wantUserID: 7,
		},
		{
			name: "invalid signature",
			cookie: &http.Cookie{
				Name:  "session",
				Value: token + ":bad-signature",
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := openAuthMiddlewareTestDB(t, token)
			var gotUser *middleware.User
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser = middleware.GetUser(r)
				w.WriteHeader(http.StatusNoContent)
			})

			h := middleware.RequireAuth(db, secret)(next)
			req := httptest.NewRequest(http.MethodGet, "/api/feed", nil)
			req.AddCookie(tc.cookie)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.wantStatus, rr.Body.String())
			}
			if tc.wantStatus == http.StatusNoContent {
				if gotUser == nil {
					t.Fatalf("expected user in context")
				}
				if gotUser.ID != tc.wantUserID {
					t.Fatalf("userID=%d want=%d", gotUser.ID, tc.wantUserID)
				}
			}
		})
	}
}

func openAuthMiddlewareTestDB(t *testing.T, token string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	stmts := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			oauth_provider TEXT NOT NULL,
			oauth_id TEXT NOT NULL,
			email TEXT,
			name TEXT,
			avatar_url TEXT,
			telegram_chat_id TEXT,
			notify_telegram BOOLEAN DEFAULT 0,
			notify_email BOOLEAN DEFAULT 1,
			is_admin BOOLEAN DEFAULT 0,
			is_active BOOLEAN DEFAULT 1,
			created_at TEXT,
			last_seen_at TEXT
		)`,
		`CREATE TABLE sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at TEXT NOT NULL,
			last_seen_at TEXT,
			created_at TEXT
		)`,
		`INSERT INTO users (
			id, oauth_provider, oauth_id, email, name, avatar_url,
			telegram_chat_id, notify_telegram, notify_email, is_admin, is_active,
			created_at, last_seen_at
		) VALUES (
			7, 'github', 'gh_7', 'user@example.com', 'Test User', 'avatar.png',
			'', 0, 1, 0, 1, '2024-01-01 00:00:00', '2024-01-01 00:00:00'
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	if _, err := db.Exec(`
		INSERT INTO sessions (token, user_id, expires_at, last_seen_at, created_at)
		VALUES (?, 7, datetime('now', '+2 day'), datetime('now'), datetime('now'))
	`, token); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	return db
}

func signSession(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}
