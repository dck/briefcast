package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

type AdminStats struct {
	Pending           int
	Processing        int
	Done              int
	Failed            int
	Skipped           int
	GroqRequestsToday int
	GroqTokensToday   int
	WorkerLastBeat    string
	RssLastRun        string
	ProcessingPaused  bool
}

type AdminEpisode struct {
	ID              int
	PodcastTitle    string
	PodcastImageURL string
	Title           string
	Status          string
	CurrentStep     string
	RetryCount      int
	LastError       string
	SkipReason      string
	PublishedAt     string
	ProcessedAt     string
	CreatedAt       string
	Logs            []AdminEpisodeLog
}

type AdminEpisodeLog struct {
	Step       string
	Status     string
	Message    string
	DurationMs int
	CreatedAt  string
}

type AdminUser struct {
	ID                int
	Email             string
	Name              string
	AvatarURL         string
	IsAdmin           bool
	IsActive          bool
	LastSeenAt        string
	CreatedAt         string
	SubscriptionCount int
}

type AdminSession struct {
	Token      string
	UserID     int
	Email      string
	Name       string
	CreatedAt  string
	LastSeenAt string
	ExpiresAt  string
}

func (r *AdminRepository) GetStats(ctx context.Context) (AdminStats, error) {
	var stats AdminStats

	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM episodes GROUP BY status`)
	if err != nil {
		return AdminStats{}, fmt.Errorf("query episode status stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return AdminStats{}, fmt.Errorf("scan episode status stats: %w", err)
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
	if err := rows.Err(); err != nil {
		return AdminStats{}, fmt.Errorf("iterate episode status stats: %w", err)
	}

	var reqCount, tokCount sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `
	SELECT COUNT(*), COALESCE(SUM(tokens_used), 0)
	FROM api_logs WHERE created_at >= date('now')`).Scan(&reqCount, &tokCount); err != nil {
		return AdminStats{}, fmt.Errorf("query groq usage stats: %w", err)
	}
	stats.GroqRequestsToday = int(reqCount.Int64)
	stats.GroqTokensToday = int(tokCount.Int64)

	var workerBeat sql.NullString
	err = r.db.QueryRowContext(ctx, `
	SELECT last_beat_at FROM worker_heartbeats
	ORDER BY last_beat_at DESC LIMIT 1`).Scan(&workerBeat)
	if err != nil && err != sql.ErrNoRows {
		return AdminStats{}, fmt.Errorf("query worker heartbeat: %w", err)
	}
	stats.WorkerLastBeat = workerBeat.String

	rssVal, err := r.getSetting(ctx, "rss_last_run")
	if err == nil {
		stats.RssLastRun = rssVal
	}

	pausedVal, err := r.getSetting(ctx, "processing_paused")
	if err == nil {
		stats.ProcessingPaused = pausedVal == "true"
	}

	return stats, nil
}

func (r *AdminRepository) ListEpisodes(ctx context.Context, statusFilter string) ([]AdminEpisode, error) {
	query := `
		SELECT e.id, COALESCE(p.title, ''), COALESCE(p.image_url, ''),
			COALESCE(e.title, ''), COALESCE(e.status, ''), COALESCE(e.current_step, ''),
			COALESCE(e.retry_count, 0), COALESCE(e.last_error, ''), COALESCE(e.skip_reason, ''),
			COALESCE(e.published_at, ''), COALESCE(e.processed_at, ''), COALESCE(e.created_at, '')
		FROM episodes e
		LEFT JOIN podcasts p ON e.podcast_id = p.id`

	var args []any
	if statusFilter != "" {
		query += " WHERE e.status = ?"
		args = append(args, statusFilter)
	}
	query += " ORDER BY e.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query admin episodes: %w", err)
	}
	defer rows.Close()

	var episodes []AdminEpisode
	idxMap := map[int]int{}

	for rows.Next() {
		var ep AdminEpisode
		if err := rows.Scan(
			&ep.ID, &ep.PodcastTitle, &ep.PodcastImageURL,
			&ep.Title, &ep.Status, &ep.CurrentStep,
			&ep.RetryCount, &ep.LastError, &ep.SkipReason,
			&ep.PublishedAt, &ep.ProcessedAt, &ep.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin episode: %w", err)
		}
		ep.Logs = []AdminEpisodeLog{}
		idxMap[ep.ID] = len(episodes)
		episodes = append(episodes, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin episodes: %w", err)
	}

	if len(episodes) > 0 {
		logRows, err := r.db.QueryContext(ctx, `
			SELECT episode_id, COALESCE(step, ''), COALESCE(status, ''),
				COALESCE(message, ''), COALESCE(duration_ms, 0), COALESCE(created_at, '')
			FROM episode_logs ORDER BY created_at ASC`)
		if err == nil {
			defer logRows.Close()
			for logRows.Next() {
				var epID int
				var l AdminEpisodeLog
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
		episodes = []AdminEpisode{}
	}
	return episodes, nil
}

func (r *AdminRepository) RetryEpisode(ctx context.Context, episodeID int) error {
	res, err := r.db.ExecContext(ctx, `UPDATE episodes SET status = 'pending', retry_count = 0 WHERE id = ?`, episodeID)
	if err != nil {
		return fmt.Errorf("retry episode: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read retry episode rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AdminRepository) RetryAllEpisode(ctx context.Context, episodeID int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE episodes SET current_step = 'download', status = 'pending',
			retry_count = 0, last_error = NULL WHERE id = ?`, episodeID)
	if err != nil {
		return fmt.Errorf("retry all episode steps: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read retry all rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AdminRepository) SkipEpisode(ctx context.Context, episodeID int, reason string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE episodes SET status = 'skipped', skip_reason = ? WHERE id = ?`, reason, episodeID)
	if err != nil {
		return fmt.Errorf("skip episode: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read skip rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AdminRepository) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, COALESCE(u.email, ''), COALESCE(u.name, ''),
			COALESCE(u.avatar_url, ''), u.is_admin, u.is_active,
			COALESCE(u.last_seen_at, ''), COALESCE(u.created_at, ''),
			COUNT(s.podcast_id) as sub_count
		FROM users u
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.active = true
		GROUP BY u.id
		ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query admin users: %w", err)
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL,
			&u.IsAdmin, &u.IsActive, &u.LastSeenAt, &u.CreatedAt,
			&u.SubscriptionCount); err != nil {
			return nil, fmt.Errorf("scan admin user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin users: %w", err)
	}
	if users == nil {
		users = []AdminUser{}
	}
	return users, nil
}

func (r *AdminRepository) DeactivateUser(ctx context.Context, userID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin deactivate user tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE users SET is_active = false WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deactivate user rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit deactivate user tx: %w", err)
	}
	return nil
}

func (r *AdminRepository) ListSessions(ctx context.Context) ([]AdminSession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.token, s.user_id, COALESCE(u.email, ''), COALESCE(u.name, ''),
			COALESCE(s.created_at, ''), COALESCE(s.last_seen_at, ''), COALESCE(s.expires_at, '')
		FROM sessions s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.expires_at > datetime('now')
		ORDER BY s.last_seen_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query admin sessions: %w", err)
	}
	defer rows.Close()

	var sessions []AdminSession
	for rows.Next() {
		var s AdminSession
		if err := rows.Scan(&s.Token, &s.UserID, &s.Email, &s.Name,
			&s.CreatedAt, &s.LastSeenAt, &s.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan admin session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin sessions: %w", err)
	}
	if sessions == nil {
		sessions = []AdminSession{}
	}
	return sessions, nil
}

func (r *AdminRepository) RevokeSession(ctx context.Context, token string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read revoke session rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AdminRepository) GetSettings(ctx context.Context) (map[string]string, error) {
	all, err := r.getAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return all, nil
}

func (r *AdminRepository) UpdateSettings(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if err := r.setSetting(ctx, key, value); err != nil {
			return fmt.Errorf("set setting %q: %w", key, err)
		}
	}
	return nil
}

func (r *AdminRepository) ResumeProcessing(ctx context.Context) error {
	if err := r.setSetting(ctx, "processing_paused", "false"); err != nil {
		return fmt.Errorf("resume processing: %w", err)
	}
	return nil
}

func (r *AdminRepository) getSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query setting %q: %w", key, err)
	}
	return value, nil
}

func (r *AdminRepository) getAllSettings(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	all := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		all[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}
	return all, nil
}

func (r *AdminRepository) setSetting(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')",
		key,
		value,
	)
	if err != nil {
		return fmt.Errorf("upsert setting: %w", err)
	}
	return nil
}
