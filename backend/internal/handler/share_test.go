package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dck/briefcast/templates"
	_ "modernc.org/sqlite" // sqlite driver for memory database
)

// ensureShareTemplateCanLoad verifies that the embedded templates filesystem
// actually contains the expected file and that SharePage constructs without
// panicking. This guards against regressions like the one that caused
// the "pattern matches no files" panic.
func TestShareTemplateCanLoad(t *testing.T) {
	// we don't need a real database for template parsing; pass nil but also
	// verify the handler construction doesn't attempt to touch the DB until
	// the returned http.HandlerFunc is invoked.
	h := SharePage(nil, templates.FS)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// TestSharePageNotFound exercises the full handler path using a small in-memory
// sqlite database to back the repository.  The schema is minimal and contains
// no rows, so every token lookup should return ErrNotFound -> 404 response.
func TestSharePageNotFound(t *testing.T) {
	// open memory-backed sqlite
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	// create the three tables referenced by the query; types are permissive
	queries := []string{
		`CREATE TABLE share_links (token TEXT, episode_id INTEGER);`,
		`CREATE TABLE episodes (id INTEGER PRIMARY KEY, title TEXT, summary TEXT, audio_url TEXT, published_at TEXT, podcast_id INTEGER, status TEXT);`,
		`CREATE TABLE podcasts (id INTEGER PRIMARY KEY, title TEXT, image_url TEXT);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/e/notoken", nil)
	r := httptest.NewRecorder()
	h := SharePage(db, templates.FS)
	h.ServeHTTP(r, req)

	if r.Code != http.StatusNotFound {
		t.Fatalf("expected 404 not found, got %d", r.Code)
	}
}
