package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dck/briefcast/internal/repository"
	"github.com/go-chi/chi/v5"
)

type StatsResponse struct {
	Pending           int    `json:"pending"`
	Processing        int    `json:"processing"`
	Done              int    `json:"done"`
	Failed            int    `json:"failed"`
	Skipped           int    `json:"skipped"`
	GroqRequestsToday int    `json:"groqRequestsToday"`
	GroqTokensToday   int    `json:"groqTokensToday"`
	WorkerLastBeat    string `json:"workerLastBeat"`
	RssLastRun        string `json:"rssLastRun"`
	ProcessingPaused  bool   `json:"processingPaused"`
}

type AdminEpisodeItem struct {
	ID              int              `json:"id"`
	PodcastTitle    string           `json:"podcastTitle"`
	PodcastImageURL string           `json:"podcastImageUrl"`
	Title           string           `json:"title"`
	Status          string           `json:"status"`
	CurrentStep     string           `json:"currentStep"`
	RetryCount      int              `json:"retryCount"`
	LastError       string           `json:"lastError"`
	SkipReason      string           `json:"skipReason"`
	PublishedAt     string           `json:"publishedAt"`
	ProcessedAt     string           `json:"processedAt"`
	CreatedAt       string           `json:"createdAt"`
	Logs            []EpisodeLogItem `json:"logs"`
}

type EpisodeLogItem struct {
	Step       string `json:"step"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	DurationMs int    `json:"durationMs"`
	CreatedAt  string `json:"createdAt"`
}

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func AdminStats(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := repo.GetStats(r.Context())
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StatsResponse{
			Pending:           stats.Pending,
			Processing:        stats.Processing,
			Done:              stats.Done,
			Failed:            stats.Failed,
			Skipped:           stats.Skipped,
			GroqRequestsToday: stats.GroqRequestsToday,
			GroqTokensToday:   stats.GroqTokensToday,
			WorkerLastBeat:    stats.WorkerLastBeat,
			RssLastRun:        stats.RssLastRun,
			ProcessingPaused:  stats.ProcessingPaused,
		})
	}
}

func AdminEpisodes(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		episodes, err := repo.ListEpisodes(r.Context(), r.URL.Query().Get("status"))
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		resp := make([]AdminEpisodeItem, 0, len(episodes))
		for _, ep := range episodes {
			item := AdminEpisodeItem{
				ID:              ep.ID,
				PodcastTitle:    ep.PodcastTitle,
				PodcastImageURL: ep.PodcastImageURL,
				Title:           ep.Title,
				Status:          ep.Status,
				CurrentStep:     ep.CurrentStep,
				RetryCount:      ep.RetryCount,
				LastError:       ep.LastError,
				SkipReason:      ep.SkipReason,
				PublishedAt:     ep.PublishedAt,
				ProcessedAt:     ep.ProcessedAt,
				CreatedAt:       ep.CreatedAt,
				Logs:            make([]EpisodeLogItem, 0, len(ep.Logs)),
			}
			for _, log := range ep.Logs {
				item.Logs = append(item.Logs, EpisodeLogItem{
					Step:       log.Step,
					Status:     log.Status,
					Message:    log.Message,
					DurationMs: log.DurationMs,
					CreatedAt:  log.CreatedAt,
				})
			}
			resp = append(resp, item)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func AdminRetryEpisode(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		err = repo.RetryEpisode(r.Context(), id)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminRetryAllEpisode(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		err = repo.RetryAllEpisode(r.Context(), id)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminSkipEpisode(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		var body struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		err = repo.SkipEpisode(r.Context(), id, body.Reason)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminUsers(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		users, err := repo.ListUsers(r.Context())
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		type userItem struct {
			ID                int    `json:"id"`
			Email             string `json:"email"`
			Name              string `json:"name"`
			AvatarURL         string `json:"avatarUrl"`
			IsAdmin           bool   `json:"isAdmin"`
			IsActive          bool   `json:"isActive"`
			LastSeenAt        string `json:"lastSeenAt"`
			CreatedAt         string `json:"createdAt"`
			SubscriptionCount int    `json:"subscriptionCount"`
		}

		resp := make([]userItem, 0, len(users))
		for _, u := range users {
			resp = append(resp, userItem{
				ID:                u.ID,
				Email:             u.Email,
				Name:              u.Name,
				AvatarURL:         u.AvatarURL,
				IsAdmin:           u.IsAdmin,
				IsActive:          u.IsActive,
				LastSeenAt:        u.LastSeenAt,
				CreatedAt:         u.CreatedAt,
				SubscriptionCount: u.SubscriptionCount,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func AdminDeactivateUser(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid user id", http.StatusBadRequest)
			return
		}

		err = repo.DeactivateUser(r.Context(), id)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminSessions(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := repo.ListSessions(r.Context())
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		type sessionItem struct {
			Token      string `json:"token"`
			UserID     int    `json:"userId"`
			Email      string `json:"email"`
			Name       string `json:"name"`
			CreatedAt  string `json:"createdAt"`
			LastSeenAt string `json:"lastSeenAt"`
			ExpiresAt  string `json:"expiresAt"`
		}

		resp := make([]sessionItem, 0, len(sessions))
		for _, s := range sessions {
			resp = append(resp, sessionItem{
				Token:      s.Token,
				UserID:     s.UserID,
				Email:      s.Email,
				Name:       s.Name,
				CreatedAt:  s.CreatedAt,
				LastSeenAt: s.LastSeenAt,
				ExpiresAt:  s.ExpiresAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func AdminRevokeSession(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			writeJSONError(w, "missing token", http.StatusBadRequest)
			return
		}

		err := repo.RevokeSession(r.Context(), token)
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, "session not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminGetSettings(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		all, err := repo.GetSettings(r.Context())
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}

func AdminUpdateSettings(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := repo.UpdateSettings(r.Context(), body); err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminResumeProcessing(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAdminRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if err := repo.ResumeProcessing(r.Context()); err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
