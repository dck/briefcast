package integration_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dck/briefcast/internal/handler"
	"github.com/dck/briefcast/internal/middleware"
	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"
)

func TestGetEpisodeMarksReadIntegration(t *testing.T) {
	t.Parallel()

	db := openEpisodeHandlerTestDB(t)

	h := handler.GetEpisode(db)
	req := httptest.NewRequest(http.MethodGet, "/api/episodes/42", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserContextKey, &middleware.User{ID: 9, IsAdmin: false})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episode_reads WHERE user_id = 9 AND episode_id = 42`).Scan(&count); err != nil {
		t.Fatalf("count reads: %v", err)
	}
	if count != 1 {
		t.Fatalf("read count = %d, want 1", count)
	}
}

func openEpisodeHandlerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	stmts := []string{
		`CREATE TABLE podcasts (id INTEGER PRIMARY KEY, title TEXT, image_url TEXT)`,
		`CREATE TABLE episodes (
			id INTEGER PRIMARY KEY,
			podcast_id INTEGER,
			title TEXT,
			description TEXT,
			audio_url TEXT,
			summary TEXT,
			status TEXT,
			published_at TEXT,
			processed_at TEXT
		)`,
		`CREATE TABLE episode_reads (
			user_id INTEGER,
			episode_id INTEGER,
			read_at TEXT,
			UNIQUE(user_id, episode_id)
		)`,
		`CREATE TABLE bookmarks (user_id INTEGER, episode_id INTEGER, created_at TEXT)`,
		`INSERT INTO podcasts (id, title, image_url) VALUES (1, 'Go Time', 'img.png')`,
		`INSERT INTO episodes (
			id, podcast_id, title, description, audio_url, summary, status, published_at, processed_at
		) VALUES (
			42, 1, 'Episode', 'Description', 'audio.mp3', 'Summary', 'done',
			'2024-01-01 00:00:00', '2024-01-01 01:00:00'
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return db
}
