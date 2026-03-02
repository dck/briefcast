package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dck/briefcast/internal/config"
	"github.com/dck/briefcast/internal/middleware"
	"github.com/dck/briefcast/internal/oauth"
	"github.com/dck/briefcast/internal/repository"
	"github.com/go-chi/chi/v5"
)

func AuthRedirect(cfg *config.Config, providers map[string]*oauth.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "provider")
		provider, ok := providers[name]
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}

		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		state := hex.EncodeToString(b)

		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_provider",
			Value:    name,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
		})

		url := provider.Config.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func AuthCallback(cfg *config.Config, db *sql.DB, providers map[string]*oauth.Provider) http.HandlerFunc {
	repo := repository.NewAuthRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, "missing state cookie", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != stateCookie.Value {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		providerCookie, err := r.Cookie("oauth_provider")
		if err != nil {
			http.Error(w, "missing provider cookie", http.StatusBadRequest)
			return
		}
		provider, ok := providers[providerCookie.Value]
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		token, err := provider.Config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			return
		}

		userInfo, err := provider.FetchUser(r.Context(), token)
		if err != nil {
			http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
			return
		}

		userID, isActive, err := repo.UpsertOAuthUser(
			r.Context(),
			providerCookie.Value,
			userInfo.ID,
			userInfo.Email,
			userInfo.Name,
			userInfo.AvatarURL,
		)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		if !isActive {
			http.Error(w, "Account deactivated", http.StatusForbidden)
			return
		}

		cookieValue, expiresAt, err := middleware.CreateSession(r.Context(), db, userID, cfg.SessionSecret)
		if err != nil {
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    cookieValue,
			Path:     "/",
			Expires:  expiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Clear OAuth cookies
		http.SetCookie(w, &http.Cookie{
			Name:   "oauth_state",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.SetCookie(w, &http.Cookie{
			Name:   "oauth_provider",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		http.Redirect(w, r, "/feed", http.StatusTemporaryRedirect)
	}
}

func Logout(db *sql.DB) http.HandlerFunc {
	repo := repository.NewAuthRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil {
			token := cookie.Value
			if i := strings.Index(token, ":"); i != -1 {
				token = token[:i]
			}
			_ = repo.DeleteSessionToken(r.Context(), token)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		w.WriteHeader(http.StatusOK)
	}
}

type meResponse struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	AvatarURL      string `json:"avatarUrl"`
	OAuthProvider  string `json:"oauthProvider"`
	TelegramChatID string `json:"telegramChatId"`
	NotifyTelegram bool   `json:"notifyTelegram"`
	NotifyEmail    bool   `json:"notifyEmail"`
	IsAdmin        bool   `json:"isAdmin"`
	CreatedAt      string `json:"createdAt"`
	LastSeenAt     string `json:"lastSeenAt"`
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.GetUser(r)
		if u == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		id := u.ID
		resp := meResponse{
			ID:             id,
			Name:           u.Name,
			Email:          u.Email,
			AvatarURL:      u.AvatarURL,
			OAuthProvider:  u.OAuthProvider,
			TelegramChatID: u.TelegramChatID,
			NotifyTelegram: u.NotifyTelegram,
			NotifyEmail:    u.NotifyEmail,
			IsAdmin:        u.IsAdmin,
			CreatedAt:      u.CreatedAt,
			LastSeenAt:     u.LastSeenAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
