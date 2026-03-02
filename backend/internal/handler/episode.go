package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"

	"github.com/dck/briefcast/internal/middleware"
	"github.com/dck/briefcast/internal/repository"
	"github.com/go-chi/chi/v5"
)

type EpisodeDetail struct {
	ID              int    `json:"id"`
	PodcastID       int    `json:"podcastId"`
	PodcastTitle    string `json:"podcastTitle"`
	PodcastImageURL string `json:"podcastImageUrl"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	AudioURL        string `json:"audioUrl"`
	Summary         string `json:"summary"`
	Status          string `json:"status"`
	PublishedAt     string `json:"publishedAt"`
	ProcessedAt     string `json:"processedAt"`
	IsRead          bool   `json:"isRead"`
	IsBookmarked    bool   `json:"isBookmarked"`
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(data []byte) string {
	intVal := new(big.Int).SetBytes(data)
	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := new(big.Int)

	var encoded []byte
	for intVal.Cmp(zero) > 0 {
		intVal.DivMod(intVal, base, mod)
		encoded = append([]byte{base58Alphabet[mod.Int64()]}, encoded...)
	}

	// Preserve leading zeros
	for _, b := range data {
		if b != 0 {
			break
		}
		encoded = append([]byte{base58Alphabet[0]}, encoded...)
	}

	return string(encoded)
}

// writeJSONError is defined in admin.go

// GetEpisode handles GET /api/episodes/{id}
func GetEpisode(db *sql.DB) http.HandlerFunc {
	repo := repository.NewEpisodeRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		episodeID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		userID := u.ID

		episode, err := repo.GetEpisodeDetail(r.Context(), userID, episodeID)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		if !u.IsAdmin && episode.Status != "done" {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}

		// Mark as read
		_ = repo.MarkRead(r.Context(), userID, episodeID)

		ep := EpisodeDetail{
			ID:              episode.ID,
			PodcastID:       episode.PodcastID,
			PodcastTitle:    episode.PodcastTitle,
			PodcastImageURL: episode.PodcastImageURL,
			Title:           episode.Title,
			Description:     episode.Description,
			AudioURL:        episode.AudioURL,
			Summary:         episode.Summary,
			Status:          episode.Status,
			PublishedAt:     episode.PublishedAt,
			ProcessedAt:     episode.ProcessedAt,
			IsRead:          true,
			IsBookmarked:    episode.IsBookmarked,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ep)
	}
}

// MarkRead handles POST /api/episodes/{id}/read
func MarkRead(db *sql.DB) http.HandlerFunc {
	repo := repository.NewEpisodeRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		episodeID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		userID := u.ID

		if err := repo.MarkRead(r.Context(), userID, episodeID); err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ToggleBookmark handles POST /api/episodes/{id}/bookmark
func ToggleBookmark(db *sql.DB) http.HandlerFunc {
	repo := repository.NewEpisodeRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		episodeID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		userID := u.ID

		bookmarked, err := repo.ToggleBookmark(r.Context(), userID, episodeID)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"bookmarked": bookmarked})
	}
}

// ShareEpisode handles POST /api/episodes/{id}/share
func ShareEpisode(db *sql.DB) http.HandlerFunc {
	repo := repository.NewShareRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		episodeID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		userID := u.ID

		// Check for existing share link
		token, err := repo.GetTokenByEpisodeID(r.Context(), episodeID)
		if errors.Is(err, repository.ErrNotFound) {
			// Generate new token
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				writeJSONError(w, "internal error", http.StatusInternalServerError)
				return
			}
			token = base58Encode(b)

			if err := repo.CreateShareLink(r.Context(), token, episodeID, userID); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"shareUrl": "/e/" + token,
			"token":    token,
		})
	}
}
