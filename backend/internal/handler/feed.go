package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dck/briefcast/internal/middleware"
	"github.com/dck/briefcast/internal/repository"
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
	repo := repository.NewFeedRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeErrorJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		page := parsePage(r)
		offset := (page - 1) * feedPageSize

		result, err := repo.ListFeed(r.Context(), u.ID, feedPageSize+1, offset)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to query feed")
			return
		}
		episodes, hasMore := trimEpisodes(result, feedPageSize)

		writeJSON(w, http.StatusOK, FeedResponse{Episodes: episodes, HasMore: hasMore})
	}
}

// Saved handles GET /api/saved?page=N
// Same as Feed but only bookmarked episodes.
func Saved(db *sql.DB) http.HandlerFunc {
	repo := repository.NewFeedRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeErrorJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		page := parsePage(r)
		offset := (page - 1) * feedPageSize

		result, err := repo.ListSaved(r.Context(), u.ID, feedPageSize+1, offset)
		if err != nil {
			writeErrorJSON(w, http.StatusInternalServerError, "failed to query saved episodes")
			return
		}
		episodes, hasMore := trimEpisodes(result, feedPageSize)

		writeJSON(w, http.StatusOK, FeedResponse{Episodes: episodes, HasMore: hasMore})
	}
}

func trimEpisodes(repoEpisodes []repository.FeedEpisode, limit int) ([]EpisodeFeedItem, bool) {
	episodes := make([]EpisodeFeedItem, 0, len(repoEpisodes))
	for _, ep := range repoEpisodes {
		episodes = append(episodes, EpisodeFeedItem{
			ID:              ep.ID,
			PodcastID:       ep.PodcastID,
			PodcastTitle:    ep.PodcastTitle,
			PodcastImageURL: ep.PodcastImageURL,
			Title:           ep.Title,
			Description:     ep.Description,
			AudioURL:        ep.AudioURL,
			PublishedAt:     ep.PublishedAt,
			IsRead:          ep.IsRead,
			IsBookmarked:    ep.IsBookmarked,
		})
	}

	hasMore := len(episodes) > limit
	if hasMore {
		episodes = episodes[:limit]
	}
	if episodes == nil {
		episodes = []EpisodeFeedItem{}
	}
	return episodes, hasMore
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
