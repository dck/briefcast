package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type FeedRepository struct {
	db *sql.DB
}

func NewFeedRepository(db *sql.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

type FeedEpisode struct {
	ID              int
	PodcastID       int
	PodcastTitle    string
	PodcastImageURL string
	Title           string
	Description     string
	AudioURL        string
	PublishedAt     string
	IsRead          bool
	IsBookmarked    bool
}

func (r *FeedRepository) ListFeed(ctx context.Context, userID, limit, offset int) ([]FeedEpisode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.podcast_id, p.title, p.image_url, e.title, e.description, e.audio_url, e.published_at,
			CASE WHEN er.user_id IS NOT NULL THEN 1 ELSE 0 END AS is_read,
			CASE WHEN b.user_id IS NOT NULL THEN 1 ELSE 0 END AS is_bookmarked
		FROM episodes e
		JOIN podcasts p ON e.podcast_id = p.id
		JOIN subscriptions s ON s.podcast_id = e.podcast_id AND s.user_id = ?
		LEFT JOIN episode_reads er ON er.episode_id = e.id AND er.user_id = ?
		LEFT JOIN bookmarks b ON b.episode_id = e.id AND b.user_id = ?
		WHERE e.status = 'done'
		ORDER BY e.published_at DESC
		LIMIT ? OFFSET ?`,
		userID, userID, userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query feed episodes: %w", err)
	}
	defer rows.Close()

	episodes, err := scanFeedEpisodes(rows)
	if err != nil {
		return nil, fmt.Errorf("scan feed episodes: %w", err)
	}
	return episodes, nil
}

func (r *FeedRepository) ListSaved(ctx context.Context, userID, limit, offset int) ([]FeedEpisode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.podcast_id, p.title, p.image_url, e.title, e.description, e.audio_url, e.published_at,
			CASE WHEN er.user_id IS NOT NULL THEN 1 ELSE 0 END AS is_read,
			1 AS is_bookmarked
		FROM episodes e
		JOIN podcasts p ON e.podcast_id = p.id
		JOIN subscriptions s ON s.podcast_id = e.podcast_id AND s.user_id = ?
		JOIN bookmarks b2 ON b2.episode_id = e.id AND b2.user_id = ?
		LEFT JOIN episode_reads er ON er.episode_id = e.id AND er.user_id = ?
		WHERE e.status = 'done'
		ORDER BY e.published_at DESC
		LIMIT ? OFFSET ?`,
		userID, userID, userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query saved episodes: %w", err)
	}
	defer rows.Close()

	episodes, err := scanFeedEpisodes(rows)
	if err != nil {
		return nil, fmt.Errorf("scan saved episodes: %w", err)
	}
	return episodes, nil
}

func scanFeedEpisodes(rows *sql.Rows) ([]FeedEpisode, error) {
	var episodes []FeedEpisode
	for rows.Next() {
		var ep FeedEpisode
		var podcastTitle, podcastImageURL, title, description, audioURL, publishedAt sql.NullString
		var isRead, isBookmarked int

		if err := rows.Scan(
			&ep.ID, &ep.PodcastID,
			&podcastTitle, &podcastImageURL,
			&title, &description, &audioURL, &publishedAt,
			&isRead, &isBookmarked,
		); err != nil {
			return nil, fmt.Errorf("scan feed episode row: %w", err)
		}

		ep.PodcastTitle = podcastTitle.String
		ep.PodcastImageURL = podcastImageURL.String
		ep.Title = title.String
		ep.Description = description.String
		ep.AudioURL = audioURL.String
		ep.PublishedAt = publishedAt.String
		ep.IsRead = isRead == 1
		ep.IsBookmarked = isBookmarked == 1

		episodes = append(episodes, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feed episodes: %w", err)
	}
	if episodes == nil {
		episodes = []FeedEpisode{}
	}
	return episodes, nil
}
