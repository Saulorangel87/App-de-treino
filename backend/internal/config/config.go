package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	AllowedOrigin string
	AppBaseURL    string
	EmailFrom     string
	ResendAPIKey  string
	SessionTTL    time.Duration
	EmailTokenTTL time.Duration
	SecureCookies bool
}

func Load() (Config, error) {
	sessionDays, err := strconv.Atoi(valueOrDefault("SESSION_DAYS", "7"))
	if err != nil || sessionDays < 1 || sessionDays > 30 {
		return Config{}, errors.New("SESSION_DAYS must be between 1 and 30")
	}
	emailTokenHours, err := strconv.Atoi(valueOrDefault("EMAIL_TOKEN_HOURS", "24"))
	if err != nil || emailTokenHours < 1 || emailTokenHours > 168 {
		return Config{}, errors.New("EMAIL_TOKEN_HOURS must be between 1 and 168")
	}
	cfg := Config{
		Port:          valueOrDefault("API_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AllowedOrigin: valueOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
		AppBaseURL:    valueOrDefault("APP_BASE_URL", "http://localhost:3000"),
		EmailFrom:     strings.TrimSpace(os.Getenv("EMAIL_FROM")),
		ResendAPIKey:  strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		SessionTTL:    time.Duration(sessionDays) * 24 * time.Hour,
		EmailTokenTTL: time.Duration(emailTokenHours) * time.Hour,
		SecureCookies: strings.EqualFold(os.Getenv("APP_ENV"), "production"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SecureCookies && (cfg.EmailFrom == "" || cfg.ResendAPIKey == "") {
		return Config{}, errors.New("EMAIL_FROM and RESEND_API_KEY are required in production")
	}
	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
