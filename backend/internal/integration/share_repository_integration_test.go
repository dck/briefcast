package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/dck/briefcast/internal/repository"
	_ "modernc.org/sqlite"
)

func TestShareRepositoryIntegration(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	stmts := []string{
		`CREATE TABLE podcasts (id INTEGER PRIMARY KEY, title TEXT, image_url TEXT)`,
		`CREATE TABLE episodes (
			id INTEGER PRIMARY KEY,
			podcast_id INTEGER NOT NULL,
			summary TEXT,
			audio_url TEXT,
			published_at TEXT,
			title TEXT,
			status TEXT
		)`,
		`CREATE TABLE share_links (
			token TEXT PRIMARY KEY,
			episode_id INTEGER UNIQUE NOT NULL,
			created_by INTEGER NOT NULL,
			created_at TEXT
		)`,
		`INSERT INTO podcasts (id, title, image_url) VALUES (1, 'Go Time', 'img.png')`,
		`INSERT INTO episodes (
			id, podcast_id, summary, audio_url, published_at, title, status
		) VALUES (
			42, 1, 'Summary text', 'audio.mp3', '2024-01-01 00:00:00', 'Episode title', 'done'
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	repo := repository.NewShareRepository(db)
	ctx := context.Background()

	if _, err := repo.GetTokenByEpisodeID(ctx, 999); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("GetTokenByEpisodeID not found error = %v, want ErrNotFound", err)
	}

	if err := repo.CreateShareLink(ctx, "tok123", 42, 7); err != nil {
		t.Fatalf("CreateShareLink error: %v", err)
	}

	token, err := repo.GetTokenByEpisodeID(ctx, 42)
	if err != nil {
		t.Fatalf("GetTokenByEpisodeID error: %v", err)
	}
	if token != "tok123" {
		t.Fatalf("token = %q, want tok123", token)
	}

	shared, err := repo.GetSharedEpisode(ctx, "tok123")
	if err != nil {
		t.Fatalf("GetSharedEpisode error: %v", err)
	}
	if shared.EpisodeTitle != "Episode title" || shared.PodcastTitle != "Go Time" || !shared.HasSummary {
		t.Fatalf("unexpected shared result: %+v", shared)
	}
}
