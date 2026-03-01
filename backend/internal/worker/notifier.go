package worker

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/briefcast/briefcast/internal/resend"
	"github.com/briefcast/briefcast/internal/telegram"
)

type Notifier struct {
	DB             *sql.DB
	TelegramClient *telegram.Client
	ResendClient   *resend.Client
	BaseURL        string
}

func NewNotifier(db *sql.DB, tgClient *telegram.Client, resendClient *resend.Client, baseURL string) *Notifier {
	return &Notifier{
		DB:             db,
		TelegramClient: tgClient,
		ResendClient:   resendClient,
		BaseURL:        baseURL,
	}
}

// NotifyForEpisode sends notifications to all subscribers of the episode's podcast.
// Called after an episode reaches status=done.
// For each subscriber with active notification preferences:
//   - Check if notification already sent (prevents duplicates on retry)
//   - Send Telegram if notify_telegram=true and telegram_chat_id is set
//   - Send email if notify_email=true and email is set
//   - Record results in notifications table
func (n *Notifier) NotifyForEpisode(episodeID int) error {
	podcastName, episodeTitle, summary, err := n.fetchEpisodeInfo(episodeID)
	if err != nil {
		return fmt.Errorf("fetch episode info: %w", err)
	}

	episodeURL, err := n.buildEpisodeURL(episodeID)
	if err != nil {
		return fmt.Errorf("build episode URL: %w", err)
	}

	teaser := extractTeaser(summary)

	rows, err := n.DB.Query(
		`SELECT u.id, u.email, u.telegram_chat_id, u.notify_telegram, u.notify_email
		 FROM users u
		 JOIN subscriptions s ON s.user_id = u.id
		 WHERE s.podcast_id = (SELECT podcast_id FROM episodes WHERE id = ?)
		   AND s.active = true
		   AND u.is_active = true`,
		episodeID,
	)
	if err != nil {
		return fmt.Errorf("query subscribers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID         int
			email          sql.NullString
			telegramChatID sql.NullString
			notifyTelegram bool
			notifyEmail    bool
		)
		if err := rows.Scan(&userID, &email, &telegramChatID, &notifyTelegram, &notifyEmail); err != nil {
			log.Printf("notifier: scan subscriber row: %v", err)
			continue
		}

		// Check if notification already sent (prevents duplicates on retry)
		var existingID int
		err := n.DB.QueryRow(
			`SELECT id FROM notifications WHERE user_id = ? AND episode_id = ?`,
			userID, episodeID,
		).Scan(&existingID)
		if err == nil {
			continue // already notified
		}
		if err != sql.ErrNoRows {
			log.Printf("notifier: check existing notification for user %d episode %d: %v", userID, episodeID, err)
			continue
		}

		var (
			sentViaTelegram bool
			sentViaEmail    bool
			telegramErr     sql.NullString
			emailErr        sql.NullString
		)

		if notifyTelegram && telegramChatID.Valid && telegramChatID.String != "" {
			msg := formatTelegramMessage(podcastName, episodeTitle, teaser, episodeURL)
			if err := n.TelegramClient.SendMessage(telegramChatID.String, msg); err != nil {
				log.Printf("notifier: telegram send to user %d: %v", userID, err)
				telegramErr = sql.NullString{String: err.Error(), Valid: true}
			} else {
				sentViaTelegram = true
			}
		}

		if notifyEmail && email.Valid && email.String != "" {
			subject := fmt.Sprintf("%s: %s — Briefcast", podcastName, episodeTitle)
			body := formatEmailBody(podcastName, episodeTitle, teaser, episodeURL)
			if err := n.ResendClient.SendEmail(email.String, subject, body); err != nil {
				log.Printf("notifier: email send to user %d: %v", userID, err)
				emailErr = sql.NullString{String: err.Error(), Valid: true}
			} else {
				sentViaEmail = true
			}
		}

		_, err = n.DB.Exec(
			`INSERT INTO notifications (user_id, episode_id, sent_via_telegram, sent_via_email, telegram_error, email_error, sent_at)
			 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			userID, episodeID, sentViaTelegram, sentViaEmail, telegramErr, emailErr,
		)
		if err != nil {
			log.Printf("notifier: insert notification record for user %d episode %d: %v", userID, episodeID, err)
		}
	}

	return rows.Err()
}

func (n *Notifier) fetchEpisodeInfo(episodeID int) (podcastName, episodeTitle, summary string, err error) {
	err = n.DB.QueryRow(
		`SELECT p.title, e.title, e.summary
		 FROM episodes e
		 JOIN podcasts p ON p.id = e.podcast_id
		 WHERE e.id = ?`,
		episodeID,
	).Scan(&podcastName, &episodeTitle, &summary)
	return
}

func (n *Notifier) buildEpisodeURL(episodeID int) (string, error) {
	var token sql.NullString
	err := n.DB.QueryRow(
		`SELECT token FROM share_links WHERE episode_id = ?`,
		episodeID,
	).Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("query share_links: %w", err)
	}
	if token.Valid && token.String != "" {
		return fmt.Sprintf("%s/e/%s", n.BaseURL, token.String), nil
	}
	return fmt.Sprintf("%s/episodes/%d", n.BaseURL, episodeID), nil
}

func extractTeaser(summary string) string {
	// First paragraph: up to first double newline
	if idx := strings.Index(summary, "\n\n"); idx != -1 {
		summary = summary[:idx]
	}
	// Trim to ~300 chars on a sentence boundary
	if len(summary) > 300 {
		cut := summary[:300]
		if lastDot := strings.LastIndex(cut, ". "); lastDot != -1 {
			return cut[:lastDot+1]
		}
		return strings.TrimSpace(cut) + "…"
	}
	return strings.TrimSpace(summary)
}

func formatTelegramMessage(podcastName, episodeTitle, teaser, url string) string {
	return fmt.Sprintf("🎙 %s\n📌 %s\n\n%s\n\nRead full summary → %s",
		podcastName, episodeTitle, teaser, url)
}

func formatEmailBody(podcastName, episodeTitle, teaser, url string) string {
	return fmt.Sprintf(`<div style="font-family: sans-serif; max-width: 560px; margin: 0 auto; padding: 24px;">
  <p style="margin: 0 0 4px; font-size: 18px; font-weight: bold;">🎙 %s</p>
  <p style="margin: 0 0 16px; font-size: 16px;">📌 %s</p>
  <p style="margin: 0 0 24px; font-size: 14px; line-height: 1.5; color: #333;">%s</p>
  <p style="margin: 0;"><a href="%s" style="color: #0066cc; font-size: 14px; text-decoration: none;">Read full summary →</a></p>
</div>`,
		podcastName, episodeTitle, teaser, url)
}
