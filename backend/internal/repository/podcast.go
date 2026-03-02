package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type PodcastRepository struct {
	db *sql.DB
}

func NewPodcastRepository(db *sql.DB) *PodcastRepository {
	return &PodcastRepository{db: db}
}

type PodcastListItem struct {
	ID           int
	Title        string
	Description  string
	ImageURL     string
	RSSURL       string
	EpisodeCount int
	Active       bool
}

type PodcastRecord struct {
	ID int
}

func (r *PodcastRepository) ListByUser(ctx context.Context, userID int) ([]PodcastListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.description, COALESCE(p.image_url, ''), p.rss_url,
			COUNT(e.id) AS episode_count, s.active
		FROM subscriptions s
		JOIN podcasts p ON p.id = s.podcast_id
		LEFT JOIN episodes e ON e.podcast_id = p.id AND e.status = 'done'
		WHERE s.user_id = $1
		GROUP BY p.id, p.title, p.description, p.image_url, p.rss_url, s.active
		ORDER BY p.title`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user podcasts: %w", err)
	}
	defer rows.Close()

	var items []PodcastListItem
	for rows.Next() {
		var item PodcastListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.ImageURL, &item.RSSURL, &item.EpisodeCount, &item.Active); err != nil {
			return nil, fmt.Errorf("scan user podcast: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user podcasts: %w", err)
	}
	if items == nil {
		items = []PodcastListItem{}
	}
	return items, nil
}

func (r *PodcastRepository) GetPodcastIDByRSSURL(ctx context.Context, rssURL string) (int, error) {
	var podcastID int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM podcasts WHERE rss_url = $1`, rssURL).Scan(&podcastID)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("query podcast by rss url: %w", err)
	}
	return podcastID, nil
}

func (r *PodcastRepository) CreatePodcast(ctx context.Context, rssURL, title, description, imageURL string) (int, error) {
	var podcastID int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO podcasts (rss_url, title, description, image_url) VALUES ($1, $2, $3, $4) RETURNING id`,
		rssURL, title, description, imageURL,
	).Scan(&podcastID)
	if err != nil {
		return 0, fmt.Errorf("insert podcast: %w", err)
	}
	return podcastID, nil
}

func (r *PodcastRepository) GetSubscriptionActive(ctx context.Context, userID, podcastID int) (bool, error) {
	var active bool
	err := r.db.QueryRowContext(ctx,
		`SELECT active FROM subscriptions WHERE user_id = $1 AND podcast_id = $2`,
		userID, podcastID,
	).Scan(&active)
	if err == sql.ErrNoRows {
		return false, ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf("query subscription: %w", err)
	}
	return active, nil
}

func (r *PodcastRepository) CreateSubscription(ctx context.Context, userID, podcastID int) error {
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (user_id, podcast_id, active) VALUES ($1, $2, true)`,
		userID, podcastID,
	); err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}
	return nil
}

func (r *PodcastRepository) ActivateSubscription(ctx context.Context, userID, podcastID int) error {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET active = true WHERE user_id = $1 AND podcast_id = $2`,
		userID, podcastID,
	); err != nil {
		return fmt.Errorf("reactivate subscription: %w", err)
	}
	return nil
}

func (r *PodcastRepository) DeactivateSubscription(ctx context.Context, userID int, podcastID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET active = false WHERE user_id = $1 AND podcast_id = $2 AND active = true`,
		userID, podcastID,
	)
	if err != nil {
		return fmt.Errorf("deactivate subscription: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deactivated rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
