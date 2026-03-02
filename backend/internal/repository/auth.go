package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

type AuthSession struct {
	UserID    int
	ExpiresAt string
}

type AuthUser struct {
	ID             int
	OAuthProvider  string
	OAuthID        string
	Email          string
	Name           string
	AvatarURL      string
	TelegramChatID string
	NotifyTelegram bool
	NotifyEmail    bool
	IsAdmin        bool
	IsActive       bool
	CreatedAt      string
	LastSeenAt     string
}

func (r *AuthRepository) UpsertOAuthUser(ctx context.Context, provider, oauthID, email, name, avatarURL string) (int, bool, error) {
	var userID int
	var isActive bool

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (oauth_provider, oauth_id, email, name, avatar_url, last_seen_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(oauth_provider, oauth_id) DO UPDATE SET
			email = excluded.email,
			name = excluded.name,
			avatar_url = excluded.avatar_url,
			last_seen_at = datetime('now')
		RETURNING id, is_active`,
		provider, oauthID, email, name, avatarURL,
	).Scan(&userID, &isActive)
	if err != nil {
		return 0, false, fmt.Errorf("upsert oauth user: %w", err)
	}

	return userID, isActive, nil
}

func (r *AuthRepository) DeleteSessionToken(ctx context.Context, token string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token); err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetSession(ctx context.Context, token string) (AuthSession, error) {
	var session AuthSession
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT user_id, expires_at FROM sessions WHERE token = ?`,
		token,
	).Scan(&session.UserID, &session.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return AuthSession{}, ErrNotFound
		}
		return AuthSession{}, fmt.Errorf("query session: %w", err)
	}
	return session, nil
}

func (r *AuthRepository) UpdateSessionExpiry(ctx context.Context, token string, expiresAt, lastSeenAt time.Time) error {
	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE sessions SET expires_at = ?, last_seen_at = ? WHERE token = ?`,
		expiresAt.Format("2006-01-02 15:04:05"),
		lastSeenAt.Format("2006-01-02 15:04:05"),
		token,
	); err != nil {
		return fmt.Errorf("update session expiry: %w", err)
	}
	return nil
}

func (r *AuthRepository) UpdateSessionLastSeen(ctx context.Context, token string, lastSeenAt time.Time) error {
	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE token = ?`,
		lastSeenAt.Format("2006-01-02 15:04:05"),
		token,
	); err != nil {
		return fmt.Errorf("update session last seen: %w", err)
	}
	return nil
}

func (r *AuthRepository) UpdateUserLastSeen(ctx context.Context, userID int, lastSeenAt time.Time) error {
	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET last_seen_at = ? WHERE id = ?`,
		lastSeenAt.Format("2006-01-02 15:04:05"),
		userID,
	); err != nil {
		return fmt.Errorf("update user last seen: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, userID int) (AuthUser, error) {
	var user AuthUser
	var email, name, avatarURL, telegramChatID, lastSeenAt sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, oauth_provider, oauth_id, email, name, avatar_url,
		       telegram_chat_id, notify_telegram, notify_email,
		       is_admin, is_active, created_at, last_seen_at
		FROM users
		WHERE id = ?`,
		userID,
	).Scan(
		&user.ID, &user.OAuthProvider, &user.OAuthID,
		&email, &name, &avatarURL,
		&telegramChatID, &user.NotifyTelegram, &user.NotifyEmail,
		&user.IsAdmin, &user.IsActive, &user.CreatedAt, &lastSeenAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return AuthUser{}, ErrNotFound
		}
		return AuthUser{}, fmt.Errorf("query user by id: %w", err)
	}

	user.Email = email.String
	user.Name = name.String
	user.AvatarURL = avatarURL.String
	user.TelegramChatID = telegramChatID.String
	user.LastSeenAt = lastSeenAt.String
	return user, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token,
		userID,
		expiresAt.Format("2006-01-02 15:04:05"),
	); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}
