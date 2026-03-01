package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	BaseURL             string
	DatabasePath        string
	AudioTmpDir         string
	SessionSecret       string
	GoogleClientID      string
	GoogleClientSecret  string
	GitHubClientID      string
	GitHubClientSecret  string
	YandexClientID      string
	YandexClientSecret  string
	GroqAPIKey          string
	ResendAPIKey        string
	ResendFromEmail     string
	TelegramBotToken    string
	TelegramAdminChatID string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error — env vars may be set directly

	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		BaseURL:             getEnv("BASE_URL", "http://localhost:5173"),
		DatabasePath:        getEnv("DATABASE_PATH", "./data/briefcast.db"),
		AudioTmpDir:         getEnv("AUDIO_TMP_DIR", "./data/audio"),
		SessionSecret:       os.Getenv("SESSION_SECRET"),
		GoogleClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:      os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:  os.Getenv("GITHUB_CLIENT_SECRET"),
		YandexClientID:      os.Getenv("YANDEX_CLIENT_ID"),
		YandexClientSecret:  os.Getenv("YANDEX_CLIENT_SECRET"),
		GroqAPIKey:          os.Getenv("GROQ_API_KEY"),
		ResendAPIKey:        os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:     getEnv("RESEND_FROM_EMAIL", "noreply@briefcast.com"),
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramAdminChatID: os.Getenv("TELEGRAM_ADMIN_CHAT_ID"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
