package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"

	"github.com/briefcast/briefcast/internal/middleware"
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

		var ep EpisodeDetail
		var description, audioURL, summary, publishedAt, processedAt, podcastImageURL sql.NullString

		err = db.QueryRowContext(r.Context(), `
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
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		if !u.IsAdmin && ep.Status != "done" {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}

		ep.PodcastImageURL = podcastImageURL.String
		ep.Description = description.String
		ep.AudioURL = audioURL.String
		ep.Summary = summary.String
		ep.PublishedAt = publishedAt.String
		ep.ProcessedAt = processedAt.String

		// Mark as read
		_, _ = db.ExecContext(r.Context(), `
			INSERT INTO episode_reads (user_id, episode_id, read_at)
			VALUES (?, ?, datetime('now'))
			ON CONFLICT(user_id, episode_id) DO UPDATE SET read_at = datetime('now')`,
			userID, episodeID,
		)
		ep.IsRead = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ep)
	}
}

// MarkRead handles POST /api/episodes/{id}/read
func MarkRead(db *sql.DB) http.HandlerFunc {
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

		_, err = db.ExecContext(r.Context(), `
			INSERT INTO episode_reads (user_id, episode_id, read_at)
			VALUES (?, ?, datetime('now'))
			ON CONFLICT(user_id, episode_id) DO UPDATE SET read_at = datetime('now')`,
			userID, episodeID,
		)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ToggleBookmark handles POST /api/episodes/{id}/bookmark
func ToggleBookmark(db *sql.DB) http.HandlerFunc {
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

		var exists bool
		err = db.QueryRowContext(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = ? AND episode_id = ?)`,
			userID, episodeID,
		).Scan(&exists)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		if exists {
			_, err = db.ExecContext(r.Context(),
				`DELETE FROM bookmarks WHERE user_id = ? AND episode_id = ?`,
				userID, episodeID,
			)
		} else {
			_, err = db.ExecContext(r.Context(),
				`INSERT INTO bookmarks (user_id, episode_id, created_at) VALUES (?, ?, datetime('now'))`,
				userID, episodeID,
			)
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"bookmarked": !exists})
	}
}

// ShareEpisode handles POST /api/episodes/{id}/share
func ShareEpisode(db *sql.DB) http.HandlerFunc {
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
		var token string
		err = db.QueryRowContext(r.Context(),
			`SELECT token FROM share_links WHERE episode_id = ?`,
			episodeID,
		).Scan(&token)

		if err == sql.ErrNoRows {
			// Generate new token
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				writeJSONError(w, "internal error", http.StatusInternalServerError)
				return
			}
			token = base58Encode(b)

			_, err = db.ExecContext(r.Context(), `
				INSERT INTO share_links (token, episode_id, created_by, created_at)
				VALUES (?, ?, ?, datetime('now'))`,
				token, episodeID, userID,
			)
			if err != nil {
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
