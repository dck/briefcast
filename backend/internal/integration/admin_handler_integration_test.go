package integration_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dck/briefcast/internal/handler"
	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"
)

func TestAdminDeactivateUserIntegration(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, is_active BOOLEAN)`); err != nil {
		t.Fatalf("create users table: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE sessions (token TEXT PRIMARY KEY, user_id INTEGER)`); err != nil {
		t.Fatalf("create sessions table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, is_active) VALUES (7, true)`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sessions (token, user_id) VALUES ('t1', 7), ('t2', 7)`); err != nil {
		t.Fatalf("insert sessions: %v", err)
	}

	h := handler.AdminDeactivateUser(db)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/7/deactivate", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "7")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}

	var isActive bool
	if err := db.QueryRow(`SELECT is_active FROM users WHERE id = 7`).Scan(&isActive); err != nil {
		t.Fatalf("query user active status: %v", err)
	}
	if isActive {
		t.Fatalf("expected user to be deactivated")
	}

	var sessionCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE user_id = 7`).Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("expected sessions deleted, got %d", sessionCount)
	}
}
