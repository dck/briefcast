package handler

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/briefcast/briefcast/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type PodcastItem struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageURL     string `json:"imageUrl"`
	RSSURL       string `json:"rssUrl"`
	EpisodeCount int    `json:"episodeCount"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Image       rssImage `xml:"image"`
	ItunesImage struct {
		Href string `xml:"href,attr"`
	} `xml:"http://www.itunes.apple.com/dtds/podcast-1.0.dtd image"`
}

type rssImage struct {
	URL string `xml:"url"`
}

func podcastJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func podcastError(w http.ResponseWriter, status int, msg string) {
	podcastJSON(w, status, map[string]string{"error": msg})
}

// ListPodcasts handles GET /api/podcasts.
// Returns the user's subscriptions with episode counts (only status='done' episodes counted).
func ListPodcasts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			podcastError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT p.id, p.title, p.description, COALESCE(p.image_url, ''), p.rss_url,
			       COUNT(e.id) AS episode_count, s.active
			FROM subscriptions s
			JOIN podcasts p ON p.id = s.podcast_id
			LEFT JOIN episodes e ON e.podcast_id = p.id AND e.status = 'done'
			WHERE s.user_id = $1
			GROUP BY p.id, p.title, p.description, p.image_url, p.rss_url, s.active
			ORDER BY p.title`, u.ID)
		if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		items := []PodcastItem{}
		for rows.Next() {
			var item PodcastItem
			var active bool
			if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.ImageURL, &item.RSSURL, &item.EpisodeCount, &active); err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !active {
				item.Title = item.Title + " (inactive)"
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		podcastJSON(w, http.StatusOK, items)
	}
}

// AddPodcast handles POST /api/podcasts.
// Accepts {"rssUrl": "https://..."}, fetches the feed, and creates or reuses the podcast and subscription.
func AddPodcast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			podcastError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var body struct {
			RSSURL string `json:"rssUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			podcastError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		rssURL := strings.TrimSpace(body.RSSURL)
		if !strings.HasPrefix(rssURL, "http://") && !strings.HasPrefix(rssURL, "https://") {
			podcastError(w, http.StatusBadRequest, "invalid RSS URL: must start with http:// or https://")
			return
		}

		// Fetch and parse the RSS feed.
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, rssURL, nil)
		if err != nil {
			podcastError(w, http.StatusBadRequest, "invalid RSS URL")
			return
		}
		req.Header.Set("User-Agent", "Briefcast/1.0")

		resp, err := client.Do(req)
		if err != nil {
			podcastError(w, http.StatusBadGateway, "failed to fetch RSS feed")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			podcastError(w, http.StatusBadGateway, "RSS feed returned non-200 status")
			return
		}

		var feed rssFeed
		if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
			podcastError(w, http.StatusUnprocessableEntity, "failed to parse RSS feed")
			return
		}

		title := strings.TrimSpace(feed.Channel.Title)
		description := strings.TrimSpace(feed.Channel.Description)
		imageURL := feed.Channel.ItunesImage.Href
		if imageURL == "" {
			imageURL = feed.Channel.Image.URL
		}

		// Upsert podcast: reuse existing row if rss_url matches.
		var podcastID int
		err = db.QueryRowContext(r.Context(),
			`SELECT id FROM podcasts WHERE rss_url = $1`, rssURL).Scan(&podcastID)
		if err == sql.ErrNoRows {
			err = db.QueryRowContext(r.Context(),
				`INSERT INTO podcasts (rss_url, title, description, image_url) VALUES ($1, $2, $3, $4) RETURNING id`,
				rssURL, title, description, imageURL).Scan(&podcastID)
			if err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
		} else if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Upsert subscription: reactivate if inactive.
		var existingActive bool
		err = db.QueryRowContext(r.Context(),
			`SELECT active FROM subscriptions WHERE user_id = $1 AND podcast_id = $2`,
			u.ID, podcastID).Scan(&existingActive)
		if err == sql.ErrNoRows {
			_, err = db.ExecContext(r.Context(),
				`INSERT INTO subscriptions (user_id, podcast_id, active) VALUES ($1, $2, true)`,
				u.ID, podcastID)
			if err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
		} else if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		} else if !existingActive {
			_, err = db.ExecContext(r.Context(),
				`UPDATE subscriptions SET active = true WHERE user_id = $1 AND podcast_id = $2`,
				u.ID, podcastID)
			if err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
		}

		item := PodcastItem{
			ID:          podcastID,
			Title:       title,
			Description: description,
			ImageURL:    imageURL,
			RSSURL:      rssURL,
		}
		podcastJSON(w, http.StatusCreated, item)
	}
}

// RemovePodcast handles DELETE /api/podcasts/{id}.
// Soft deletes the subscription by setting active=false.
func RemovePodcast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			podcastError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		podcastID := chi.URLParam(r, "id")
		if podcastID == "" {
			podcastError(w, http.StatusBadRequest, "missing podcast id")
			return
		}

		res, err := db.ExecContext(r.Context(),
			`UPDATE subscriptions SET active = false WHERE user_id = $1 AND podcast_id = $2 AND active = true`,
			u.ID, podcastID)
		if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		n, err := res.RowsAffected()
		if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if n == 0 {
			podcastError(w, http.StatusNotFound, "subscription not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
