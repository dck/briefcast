package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/briefcast/briefcast/internal/settings"
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
	return func(w http.ResponseWriter, r *http.Request) {
		var stats StatsResponse

		rows, err := db.QueryContext(r.Context(), `
			SELECT status, COUNT(*) FROM episodes GROUP BY status`)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
			switch status {
			case "pending":
				stats.Pending = count
			case "processing":
				stats.Processing = count
			case "done":
				stats.Done = count
			case "failed":
				stats.Failed = count
			case "skipped":
				stats.Skipped = count
			}
		}

		var reqCount, tokCount sql.NullInt64
		err = db.QueryRowContext(r.Context(), `
			SELECT COUNT(*), COALESCE(SUM(tokens_used), 0)
			FROM api_logs WHERE created_at >= date('now')`).Scan(&reqCount, &tokCount)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		stats.GroqRequestsToday = int(reqCount.Int64)
		stats.GroqTokensToday = int(tokCount.Int64)

		var workerBeat sql.NullString
		err = db.QueryRowContext(r.Context(), `
			SELECT last_beat_at FROM worker_heartbeats
			ORDER BY last_beat_at DESC LIMIT 1`).Scan(&workerBeat)
		if err != nil && err != sql.ErrNoRows {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		stats.WorkerLastBeat = workerBeat.String

		rssVal, _ := settings.Get(db, "rss_last_run")
		stats.RssLastRun = rssVal

		pausedVal, _ := settings.Get(db, "processing_paused")
		stats.ProcessingPaused = pausedVal == "true"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

func AdminEpisodes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT e.id, COALESCE(p.title, ''), COALESCE(p.image_url, ''),
			       COALESCE(e.title, ''), COALESCE(e.status, ''), COALESCE(e.current_step, ''),
			       COALESCE(e.retry_count, 0), COALESCE(e.last_error, ''), COALESCE(e.skip_reason, ''),
			       COALESCE(e.published_at, ''), COALESCE(e.processed_at, ''), COALESCE(e.created_at, '')
			FROM episodes e
			LEFT JOIN podcasts p ON e.podcast_id = p.id`

		statusFilter := r.URL.Query().Get("status")
		var args []interface{}
		if statusFilter != "" {
			query += " WHERE e.status = ?"
			args = append(args, statusFilter)
		}
		query += " ORDER BY e.created_at DESC"

		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var episodes []AdminEpisodeItem
		var episodeIDs []int
		for rows.Next() {
			var ep AdminEpisodeItem
			if err := rows.Scan(&ep.ID, &ep.PodcastTitle, &ep.PodcastImageURL,
				&ep.Title, &ep.Status, &ep.CurrentStep,
				&ep.RetryCount, &ep.LastError, &ep.SkipReason,
				&ep.PublishedAt, &ep.ProcessedAt, &ep.CreatedAt); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
			ep.Logs = []EpisodeLogItem{}
			episodes = append(episodes, ep)
			episodeIDs = append(episodeIDs, ep.ID)
		}

		// Build a map for fast lookup
		idxMap := make(map[int]int, len(episodes))
		for i, ep := range episodes {
			idxMap[ep.ID] = i
		}

		if len(episodeIDs) > 0 {
			logRows, err := db.QueryContext(r.Context(), `
				SELECT episode_id, COALESCE(step, ''), COALESCE(status, ''),
				       COALESCE(message, ''), COALESCE(duration_ms, 0), COALESCE(created_at, '')
				FROM episode_logs ORDER BY created_at ASC`)
			if err == nil {
				defer logRows.Close()
				for logRows.Next() {
					var epID int
					var l EpisodeLogItem
					if err := logRows.Scan(&epID, &l.Step, &l.Status, &l.Message, &l.DurationMs, &l.CreatedAt); err != nil {
						continue
					}
					if idx, ok := idxMap[epID]; ok {
						episodes[idx].Logs = append(episodes[idx].Logs, l)
					}
				}
			}
		}

		if episodes == nil {
			episodes = []AdminEpisodeItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(episodes)
	}
}

func AdminRetryEpisode(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		res, err := db.ExecContext(r.Context(), `
			UPDATE episodes SET status = 'pending', retry_count = 0 WHERE id = ?`, id)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminRetryAllEpisode(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid episode id", http.StatusBadRequest)
			return
		}

		res, err := db.ExecContext(r.Context(), `
			UPDATE episodes SET current_step = 'download', status = 'pending',
			       retry_count = 0, last_error = NULL WHERE id = ?`, id)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminSkipEpisode(db *sql.DB) http.HandlerFunc {
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

		res, err := db.ExecContext(r.Context(), `
			UPDATE episodes SET status = 'skipped', skip_reason = ? WHERE id = ?`, body.Reason, id)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSONError(w, "episode not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), `
			SELECT u.id, COALESCE(u.email, ''), COALESCE(u.name, ''),
			       COALESCE(u.avatar_url, ''), u.is_admin, u.is_active,
			       COALESCE(u.last_seen_at, ''), COALESCE(u.created_at, ''),
			       COUNT(s.podcast_id) as sub_count
			FROM users u
			LEFT JOIN subscriptions s ON u.id = s.user_id AND s.active = true
			GROUP BY u.id
			ORDER BY u.created_at DESC`)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

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

		var users []userItem
		for rows.Next() {
			var u userItem
			if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL,
				&u.IsAdmin, &u.IsActive, &u.LastSeenAt, &u.CreatedAt,
				&u.SubscriptionCount); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		if users == nil {
			users = []userItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func AdminDeactivateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeJSONError(w, "invalid user id", http.StatusBadRequest)
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		res, err := tx.ExecContext(r.Context(), `UPDATE users SET is_active = false WHERE id = ?`, id)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSONError(w, "user not found", http.StatusNotFound)
			return
		}

		_, err = tx.ExecContext(r.Context(), `DELETE FROM sessions WHERE user_id = ?`, id)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminSessions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), `
			SELECT s.token, s.user_id, COALESCE(u.email, ''), COALESCE(u.name, ''),
			       COALESCE(s.created_at, ''), COALESCE(s.last_seen_at, ''), COALESCE(s.expires_at, '')
			FROM sessions s
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.expires_at > datetime('now')
			ORDER BY s.last_seen_at DESC`)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type sessionItem struct {
			Token      string `json:"token"`
			UserID     int    `json:"userId"`
			Email      string `json:"email"`
			Name       string `json:"name"`
			CreatedAt  string `json:"createdAt"`
			LastSeenAt string `json:"lastSeenAt"`
			ExpiresAt  string `json:"expiresAt"`
		}

		var sessions []sessionItem
		for rows.Next() {
			var s sessionItem
			if err := rows.Scan(&s.Token, &s.UserID, &s.Email, &s.Name,
				&s.CreatedAt, &s.LastSeenAt, &s.ExpiresAt); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
			sessions = append(sessions, s)
		}

		if sessions == nil {
			sessions = []sessionItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	}
}

func AdminRevokeSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			writeJSONError(w, "missing token", http.StatusBadRequest)
			return
		}

		res, err := db.ExecContext(r.Context(), `DELETE FROM sessions WHERE token = ?`, token)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			writeJSONError(w, "session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminGetSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := settings.GetAll(db)
		if err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}

func AdminUpdateSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		for key, value := range body {
			if err := settings.Set(db, key, value); err != nil {
				writeJSONError(w, "database error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func AdminResumeProcessing(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := settings.Set(db, "processing_paused", "false"); err != nil {
			writeJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
