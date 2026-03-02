package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type EpisodeRepository struct {
	db *sql.DB
}

func NewEpisodeRepository(db *sql.DB) *EpisodeRepository {
	return &EpisodeRepository{db: db}
}

type EpisodeDetail struct {
	ID              int
	PodcastID       int
	PodcastTitle    string
	PodcastImageURL string
	Title           string
	Description     string
	AudioURL        string
	Summary         string
	Status          string
	PublishedAt     string
	ProcessedAt     string
	IsRead          bool
	IsBookmarked    bool
}

func (r *EpisodeRepository) GetEpisodeDetail(ctx context.Context, userID, episodeID int) (EpisodeDetail, error) {
	var ep EpisodeDetail
	var description, audioURL, summary, publishedAt, processedAt, podcastImageURL sql.NullString

	err := r.db.QueryRowContext(ctx, `
	SELECT e.id, e.podcast_id, p.title, p.image_url,
e.title, e.description, e.audio_url, e.summary, e.status,
e.published_at, e.processed_at,
CASE WHEN er.user_id IS NOT NULL THEN 1 ELSE 0 END,
CASE WHEN b.user_id IS NOT NULL THEN 1 ELSE 0 END
	FROM episodes e
	JOIN podcasts p ON p.id = e.podcast_id
	LEFT JOIN episode_reads er ON er.episode_id = e.id AND er.user_id = ?
	LEFT JOIN bookmarks b ON b.episode_id = e.id AND b.user_id = ?
	WHERE e.id = ?`,
		userID, userID, episodeID,
	).Scan(
		&ep.ID, &ep.PodcastID, &ep.PodcastTitle, &podcastImageURL,
		&ep.Title, &description, &audioURL, &summary, &ep.Status,
		&publishedAt, &processedAt,
		&ep.IsRead, &ep.IsBookmarked,
	)
	if err == sql.ErrNoRows {
		return EpisodeDetail{}, ErrNotFound
	}
	if err != nil {
		return EpisodeDetail{}, fmt.Errorf("query episode detail: %w", err)
	}

	ep.PodcastImageURL = podcastImageURL.String
	ep.Description = description.String
	ep.AudioURL = audioURL.String
	ep.Summary = summary.String
	ep.PublishedAt = publishedAt.String
	ep.ProcessedAt = processedAt.String

	return ep, nil
}

func (r *EpisodeRepository) MarkRead(ctx context.Context, userID, episodeID int) error {
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO episode_reads (user_id, episode_id, read_at)
	VALUES (?, ?, datetime('now'))
	ON CONFLICT(user_id, episode_id) DO UPDATE SET read_at = datetime('now')`,
		userID, episodeID,
	)
	if err != nil {
		return fmt.Errorf("mark episode as read: %w", err)
	}
	return nil
}

func (r *EpisodeRepository) ToggleBookmark(ctx context.Context, userID, episodeID int) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = ? AND episode_id = ?)`,
		userID, episodeID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check bookmark existence: %w", err)
	}

	if exists {
		if _, err := r.db.ExecContext(ctx,
			`DELETE FROM bookmarks WHERE user_id = ? AND episode_id = ?`,
			userID, episodeID,
		); err != nil {
			return false, fmt.Errorf("delete bookmark: %w", err)
		}
		return false, nil
	}

	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO bookmarks (user_id, episode_id, created_at) VALUES (?, ?, datetime('now'))`,
		userID, episodeID,
	); err != nil {
		return false, fmt.Errorf("create bookmark: %w", err)
	}
	return true, nil
}
