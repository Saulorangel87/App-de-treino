package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port             string
	DatabaseURL      string
	AllowedOrigin    string
	AppBaseURL       string
	EmailFrom        string
	ResendAPIKey     string
	FeedbackDigestTo string
	SessionTTL       time.Duration
	EmailTokenTTL    time.Duration
	SecureCookies    bool
	AIEnabled        bool
	AIProvider       string
	AIBaseURL        string
	AIModel          string
	AIWorkerURL      string
	AIWorkerToken    string
	AITimeout        time.Duration
	AIMaxTokens      int
	AIMaxConcurrent  int
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
	aiEnabled, err := strconv.ParseBool(valueOrDefault("AI_ENABLED", "false"))
	if err != nil {
		return Config{}, errors.New("AI_ENABLED must be true or false")
	}
	aiTimeoutSeconds, err := strconv.Atoi(valueOrDefault("AI_TIMEOUT_SECONDS", "15"))
	if err != nil || aiTimeoutSeconds < 1 || aiTimeoutSeconds > 60 {
		return Config{}, errors.New("AI_TIMEOUT_SECONDS must be between 1 and 60")
	}
	aiMaxTokens, err := strconv.Atoi(valueOrDefault("AI_MAX_OUTPUT_TOKENS", "220"))
	if err != nil || aiMaxTokens < 32 || aiMaxTokens > 512 {
		return Config{}, errors.New("AI_MAX_OUTPUT_TOKENS must be between 32 and 512")
	}
	aiMaxConcurrent, err := strconv.Atoi(valueOrDefault("AI_MAX_CONCURRENT", "1"))
	if err != nil || aiMaxConcurrent < 1 || aiMaxConcurrent > 2 {
		return Config{}, errors.New("AI_MAX_CONCURRENT must be between 1 and 2")
	}
	cfg := Config{
		Port:             valueOrDefault("API_PORT", "8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		AllowedOrigin:    valueOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
		AppBaseURL:       valueOrDefault("APP_BASE_URL", "http://localhost:3000"),
		EmailFrom:        strings.TrimSpace(os.Getenv("EMAIL_FROM")),
		ResendAPIKey:     strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		FeedbackDigestTo: strings.TrimSpace(os.Getenv("FEEDBACK_DIGEST_TO")),
		SessionTTL:       time.Duration(sessionDays) * 24 * time.Hour,
		EmailTokenTTL:    time.Duration(emailTokenHours) * time.Hour,
		SecureCookies:    strings.EqualFold(os.Getenv("APP_ENV"), "production"),
		AIEnabled:        aiEnabled,
		AIProvider:       valueOrDefault("AI_PROVIDER", "ollama"),
		AIBaseURL:        valueOrDefault("AI_BASE_URL", "http://127.0.0.1:11434"),
		AIModel:          valueOrDefault("AI_MODEL", "qwen3:4b-instruct"),
		AIWorkerURL:      strings.TrimRight(strings.TrimSpace(os.Getenv("AI_WORKER_URL")), "/"),
		AIWorkerToken:    strings.TrimSpace(os.Getenv("AI_WORKER_TOKEN")),
		AITimeout:        time.Duration(aiTimeoutSeconds) * time.Second,
		AIMaxTokens:      aiMaxTokens,
		AIMaxConcurrent:  aiMaxConcurrent,
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
