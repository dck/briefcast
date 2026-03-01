package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/briefcast/briefcast/internal/middleware"
)

type UserSettings struct {
	Email          string `json:"email"`
	TelegramChatID string `json:"telegramChatId"`
	NotifyTelegram bool   `json:"notifyTelegram"`
	NotifyEmail    bool   `json:"notifyEmail"`
}

// GetSettings handles GET /api/settings
// Returns current user's notification preferences.
func GetSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		resp := UserSettings{
			Email:          u.Email,
			TelegramChatID: u.TelegramChatID,
			NotifyTelegram: u.NotifyTelegram,
			NotifyEmail:    u.NotifyEmail,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// UpdateSettings handles PUT /api/settings
// Updates email, telegram_chat_id, notify_telegram, notify_email for the current user.
func UpdateSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		var settings UserSettings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		_, err := db.ExecContext(r.Context(),
			`UPDATE users SET email = ?, telegram_chat_id = ?, notify_telegram = ?, notify_email = ? WHERE id = ?`,
			settings.Email, settings.TelegramChatID, settings.NotifyTelegram, settings.NotifyEmail, u.ID,
		)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to update settings"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
	}
}
