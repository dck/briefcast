package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/briefcast/briefcast/internal/middleware"
)

type EpisodeFeedItem struct {
	ID              int    `json:"id"`
	PodcastID       int    `json:"podcastId"`
	PodcastTitle    string `json:"podcastTitle"`
	PodcastImageURL string `json:"podcastImageUrl"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	AudioURL        string `json:"audioUrl"`
	PublishedAt     string `json:"publishedAt"`
	IsRead          bool   `json:"isRead"`
	IsBookmarked    bool   `json:"isBookmarked"`
}

type FeedResponse struct {
	Episodes []EpisodeFeedItem `json:"episodes"`
	HasMore  bool              `json:"hasMore"`
}

const feedPageSize = 20

// Feed handles GET /api/feed?page=N
// Returns paginated episodes from user's subscriptions where status='done'.
// 20 per page. Sorted by published_at DESC.
// Includes is_read and is_bookmarked per user.
func Feed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeErrorJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		page := parsePage(r)
		offset := (page - 1) * feedPageSize

		query := `
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
			LIMIT ? OFFSET ?`

		rows, err := db.Query(query, u.ID, u.ID, u.ID, feedPageSize+1, offset)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to query feed")
			return
		}
		defer rows.Close()

		episodes, hasMore, err := scanEpisodes(rows, feedPageSize)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to read feed")
			return
		}

		writeJSON(w, http.StatusOK, FeedResponse{Episodes: episodes, HasMore: hasMore})
	}
}

// Saved handles GET /api/saved?page=N
// Same as Feed but only bookmarked episodes.
func Saved(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeErrorJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		page := parsePage(r)
		offset := (page - 1) * feedPageSize

		query := `
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
			LIMIT ? OFFSET ?`

		rows, err := db.Query(query, u.ID, u.ID, u.ID, feedPageSize+1, offset)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to query saved episodes")
			return
		}
		defer rows.Close()

		episodes, hasMore, err := scanEpisodes(rows, feedPageSize)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to read saved episodes")
			return
		}

		writeJSON(w, http.StatusOK, FeedResponse{Episodes: episodes, HasMore: hasMore})
	}
}

func scanEpisodes(rows *sql.Rows, limit int) ([]EpisodeFeedItem, bool, error) {
	var episodes []EpisodeFeedItem
	for rows.Next() {
		var ep EpisodeFeedItem
		var podcastTitle, podcastImageURL, title, description, audioURL, publishedAt sql.NullString
		var isRead, isBookmarked int

		err := rows.Scan(
			&ep.ID, &ep.PodcastID,
			&podcastTitle, &podcastImageURL,
			&title, &description, &audioURL, &publishedAt,
			&isRead, &isBookmarked,
		)
		if err != nil {
			return nil, false, err
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
		return nil, false, err
	}

	hasMore := len(episodes) > limit
	if hasMore {
		episodes = episodes[:limit]
	}
	if episodes == nil {
		episodes = []EpisodeFeedItem{}
	}
	return episodes, hasMore, nil
}

func parsePage(r *http.Request) int {
	p, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || p < 1 {
		return 1
	}
	return p
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErrorJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
