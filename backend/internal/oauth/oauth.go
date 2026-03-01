package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/briefcast/briefcast/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Provider struct {
	Config  *oauth2.Config
	Name    string
	UserURL string // URL to fetch user profile info
}

type UserInfo struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

// Providers returns a map of provider name → Provider config.
// Callback URL is {BaseURL}/api/auth/callback?provider={name}
func Providers(cfg *config.Config) map[string]*Provider {
	providers := make(map[string]*Provider)
	redirectURL := cfg.BaseURL + "/api/auth/callback"

	// Google
	if cfg.GoogleClientID != "" {
		providers["google"] = &Provider{
			Name:    "google",
			UserURL: "https://www.googleapis.com/oauth2/v2/userinfo",
			Config: &oauth2.Config{
				ClientID:     cfg.GoogleClientID,
				ClientSecret: cfg.GoogleClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint:     google.Endpoint,
			},
		}
	}

	// GitHub
	if cfg.GitHubClientID != "" {
		providers["github"] = &Provider{
			Name:    "github",
			UserURL: "https://api.github.com/user",
			Config: &oauth2.Config{
				ClientID:     cfg.GitHubClientID,
				ClientSecret: cfg.GitHubClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"user:email", "read:user"},
				Endpoint:     github.Endpoint,
			},
		}
	}

	// Yandex
	if cfg.YandexClientID != "" {
		providers["yandex"] = &Provider{
			Name:    "yandex",
			UserURL: "https://login.yandex.ru/info?format=json",
			Config: &oauth2.Config{
				ClientID:     cfg.YandexClientID,
				ClientSecret: cfg.YandexClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"login:email", "login:info", "login:avatar"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://oauth.yandex.com/authorize",
					TokenURL: "https://oauth.yandex.com/token",
				},
			},
		}
	}

	return providers
}

// FetchUser fetches user info from the provider using the given token.
func (p *Provider) FetchUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.Config.Client(ctx, token)
	resp, err := client.Get(p.UserURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo *UserInfo

	switch p.Name {
	case "google":
		var data struct {
			ID      string `json:"id"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse Google user info: %w", err)
		}
		userInfo = &UserInfo{
			ID:        data.ID,
			Email:     data.Email,
			Name:      data.Name,
			AvatarURL: data.Picture,
		}

	case "github":
		var data struct {
			ID        int    `json:"id"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse GitHub user info: %w", err)
		}
		name := data.Name
		if name == "" {
			name = data.Login
		}
		userInfo = &UserInfo{
			ID:        fmt.Sprintf("%d", data.ID),
			Email:     data.Email,
			Name:      name,
			AvatarURL: data.AvatarURL,
		}

	case "yandex":
		var data struct {
			ID              string `json:"id"`
			DefaultEmail    string `json:"default_email"`
			DisplayName     string `json:"display_name"`
			DefaultAvatarID string `json:"default_avatar_id"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse Yandex user info: %w", err)
		}
		avatarURL := ""
		if data.DefaultAvatarID != "" {
			avatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", data.DefaultAvatarID)
		}
		userInfo = &UserInfo{
			ID:        data.ID,
			Email:     data.DefaultEmail,
			Name:      data.DisplayName,
			AvatarURL: avatarURL,
		}
	}

	return userInfo, nil
}
