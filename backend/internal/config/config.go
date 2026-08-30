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
	SessionTTL    time.Duration
	SecureCookies bool
}

func Load() (Config, error) {
	sessionDays, err := strconv.Atoi(valueOrDefault("SESSION_DAYS", "7"))
	if err != nil || sessionDays < 1 || sessionDays > 30 {
		return Config{}, errors.New("SESSION_DAYS must be between 1 and 30")
	}
	cfg := Config{
		Port:          valueOrDefault("API_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AllowedOrigin: valueOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
		SessionTTL:    time.Duration(sessionDays) * 24 * time.Hour,
		SecureCookies: strings.EqualFold(os.Getenv("APP_ENV"), "production"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
