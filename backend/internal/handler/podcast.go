package handler

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dck/briefcast/internal/middleware"
	"github.com/dck/briefcast/internal/repository"
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
	repo := repository.NewPodcastRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			podcastError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		podcasts, err := repo.ListByUser(r.Context(), u.ID)
		if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		items := make([]PodcastItem, 0, len(podcasts))
		for _, p := range podcasts {
			item := PodcastItem{
				ID:           p.ID,
				Title:        p.Title,
				Description:  p.Description,
				ImageURL:     p.ImageURL,
				RSSURL:       p.RSSURL,
				EpisodeCount: p.EpisodeCount,
			}
			if !p.Active {
				item.Title = item.Title + " (inactive)"
			}
			items = append(items, item)
		}

		podcastJSON(w, http.StatusOK, items)
	}
}

// AddPodcast handles POST /api/podcasts.
// Accepts {"rssUrl": "https://..."}, fetches the feed, and creates or reuses the podcast and subscription.
func AddPodcast(db *sql.DB) http.HandlerFunc {
	repo := repository.NewPodcastRepository(db)

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
		podcastID, err := repo.GetPodcastIDByRSSURL(r.Context(), rssURL)
		if errors.Is(err, repository.ErrNotFound) {
			podcastID, err = repo.CreatePodcast(r.Context(), rssURL, title, description, imageURL)
			if err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
		} else if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Upsert subscription: reactivate if inactive.
		existingActive, err := repo.GetSubscriptionActive(r.Context(), u.ID, podcastID)
		if errors.Is(err, repository.ErrNotFound) {
			if err := repo.CreateSubscription(r.Context(), u.ID, podcastID); err != nil {
				podcastError(w, http.StatusInternalServerError, "internal error")
				return
			}
		} else if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		} else if !existingActive {
			if err := repo.ActivateSubscription(r.Context(), u.ID, podcastID); err != nil {
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
	repo := repository.NewPodcastRepository(db)

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

		if err := repo.DeactivateSubscription(r.Context(), u.ID, podcastID); errors.Is(err, repository.ErrNotFound) {
			podcastError(w, http.StatusNotFound, "subscription not found")
			return
		} else if err != nil {
			podcastError(w, http.StatusInternalServerError, "internal error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
