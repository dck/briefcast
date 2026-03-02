package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) UpdateUserSettings(ctx context.Context, userID int, email, telegramChatID string, notifyTelegram, notifyEmail bool) error {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = ?, telegram_chat_id = ?, notify_telegram = ?, notify_email = ? WHERE id = ?`,
		email, telegramChatID, notifyTelegram, notifyEmail, userID,
	); err != nil {
		return fmt.Errorf("update user settings: %w", err)
	}
	return nil
}
