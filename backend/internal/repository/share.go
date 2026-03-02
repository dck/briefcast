package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ShareRepository struct {
	db *sql.DB
}

func NewShareRepository(db *sql.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

type SharedEpisode struct {
	EpisodeTitle    string
	PodcastTitle    string
	PodcastImageURL string
	Summary         string
	HasSummary      bool
	PublishedAt     string
	AudioURL        string
}

func (r *ShareRepository) GetTokenByEpisodeID(ctx context.Context, episodeID int) (string, error) {
	var token string
	err := r.db.QueryRowContext(ctx,
		`SELECT token FROM share_links WHERE episode_id = ?`,
		episodeID,
	).Scan(&token)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query share token by episode id: %w", err)
	}
	return token, nil
}

func (r *ShareRepository) CreateShareLink(ctx context.Context, token string, episodeID, createdBy int) error {
	if _, err := r.db.ExecContext(ctx, `
	INSERT INTO share_links (token, episode_id, created_by, created_at)
	VALUES (?, ?, ?, datetime('now'))`,
		token, episodeID, createdBy,
	); err != nil {
		return fmt.Errorf("insert share link: %w", err)
	}
	return nil
}

func (r *ShareRepository) GetSharedEpisode(ctx context.Context, token string) (SharedEpisode, error) {
	var shared SharedEpisode
	var summary sql.NullString
	var audioURL, imageURL, publishedAt sql.NullString

	err := r.db.QueryRowContext(ctx, `
	SELECT e.title, p.title, p.image_url, e.summary, e.audio_url, e.published_at
	FROM share_links sl
	JOIN episodes e ON sl.episode_id = e.id
	JOIN podcasts p ON e.podcast_id = p.id
	WHERE sl.token = ? AND e.status = 'done'
`, token).Scan(
		&shared.EpisodeTitle,
		&shared.PodcastTitle,
		&imageURL,
		&summary,
		&audioURL,
		&publishedAt,
	)
	if err == sql.ErrNoRows {
		return SharedEpisode{}, ErrNotFound
	}
	if err != nil {
		return SharedEpisode{}, fmt.Errorf("query shared episode by token: %w", err)
	}

	shared.PodcastImageURL = imageURL.String
	shared.AudioURL = audioURL.String
	shared.PublishedAt = publishedAt.String
	shared.HasSummary = summary.Valid
	shared.Summary = summary.String

	return shared, nil
}
