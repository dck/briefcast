package integration_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dck/briefcast/internal/repository"
	_ "modernc.org/sqlite"
)

func TestFeedRepositoryListFeedIntegration(t *testing.T) {
	t.Parallel()

	db := openFeedRepositoryTestDB(t)
	repo := repository.NewFeedRepository(db)

	episodes, err := repo.ListFeed(context.Background(), 1, 21, 0)
	if err != nil {
		t.Fatalf("ListFeed() error = %v", err)
	}

	if len(episodes) != 2 {
		t.Fatalf("len(episodes) = %d, want 2", len(episodes))
	}
	if episodes[0].ID != 102 {
		t.Fatalf("episodes[0].ID = %d, want 102", episodes[0].ID)
	}
	if episodes[1].ID != 101 {
		t.Fatalf("episodes[1].ID = %d, want 101", episodes[1].ID)
	}
	if !episodes[0].IsBookmarked {
		t.Fatalf("episodes[0].IsBookmarked = false, want true")
	}
	if !episodes[1].IsRead {
		t.Fatalf("episodes[1].IsRead = false, want true")
	}
}

func openFeedRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	stmts := []string{
		`CREATE TABLE podcasts (id INTEGER PRIMARY KEY, title TEXT, image_url TEXT)`,
		`CREATE TABLE subscriptions (user_id INTEGER, podcast_id INTEGER)`,
		`CREATE TABLE episodes (
			id INTEGER PRIMARY KEY,
			podcast_id INTEGER,
			title TEXT,
			description TEXT,
			audio_url TEXT,
			published_at TEXT,
			status TEXT
		)`,
		`CREATE TABLE episode_reads (user_id INTEGER, episode_id INTEGER)`,
		`CREATE TABLE bookmarks (user_id INTEGER, episode_id INTEGER)`,
		`INSERT INTO podcasts (id, title, image_url) VALUES (1, 'Go Time', 'img.png')`,
		`INSERT INTO subscriptions (user_id, podcast_id) VALUES (1, 1)`,
		`INSERT INTO episodes (
			id, podcast_id, title, description, audio_url, published_at, status
		) VALUES
			(101, 1, 'E1', 'Desc1', 'a1.mp3', '2024-01-01 10:00:00', 'done'),
			(102, 1, 'E2', 'Desc2', 'a2.mp3', '2024-01-02 10:00:00', 'done'),
			(103, 1, 'E3', 'Desc3', 'a3.mp3', '2024-01-03 10:00:00', 'processing')`,
		`INSERT INTO episode_reads (user_id, episode_id) VALUES (1, 101)`,
		`INSERT INTO bookmarks (user_id, episode_id) VALUES (1, 102)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}
	return db
}
