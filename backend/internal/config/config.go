package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
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
	// Development only enables responses that expose e-mail action links. It is
	// fail-closed: it requires a non-production APP_ENV and loopback URLs.
	Development     bool
	DBMaxConns      int32
	DBMinConns      int32
	AIEnabled       bool
	AIProvider      string
	AIBaseURL       string
	AIModel         string
	AIWorkerURL     string
	AIWorkerToken   string
	AITimeout       time.Duration
	AIMaxTokens     int
	AIMaxConcurrent int
	// ProtectionLevelsEnabled switches prescription from the legacy all-or-nothing
	// recovery rule to graduated, dated protection levels.
	ProtectionLevelsEnabled bool
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
	protectionLevels, err := strconv.ParseBool(valueOrDefault("PROTECTION_LEVELS_ENABLED", "false"))
	if err != nil {
		return Config{}, errors.New("PROTECTION_LEVELS_ENABLED must be true or false")
	}
	dbMaxConns, err := strconv.ParseInt(valueOrDefault("DB_MAX_CONNS", "10"), 10, 32)
	if err != nil || dbMaxConns < 1 || dbMaxConns > 50 {
		return Config{}, errors.New("DB_MAX_CONNS must be between 1 and 50")
	}
	dbMinConns, err := strconv.ParseInt(valueOrDefault("DB_MIN_CONNS", "1"), 10, 32)
	if err != nil || dbMinConns < 0 || dbMinConns > dbMaxConns {
		return Config{}, errors.New("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS")
	}
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	switch appEnv {
	case "", "development", "test", "production":
	default:
		return Config{}, fmt.Errorf("APP_ENV must be development, test or production, got %q", appEnv)
	}
	cfg := Config{
		Port:                    valueOrDefault("API_PORT", "8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		AllowedOrigin:           valueOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
		AppBaseURL:              valueOrDefault("APP_BASE_URL", "http://localhost:3000"),
		EmailFrom:               strings.TrimSpace(os.Getenv("EMAIL_FROM")),
		ResendAPIKey:            strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		FeedbackDigestTo:        strings.TrimSpace(os.Getenv("FEEDBACK_DIGEST_TO")),
		SessionTTL:              time.Duration(sessionDays) * 24 * time.Hour,
		EmailTokenTTL:           time.Duration(emailTokenHours) * time.Hour,
		SecureCookies:           appEnv == "production",
		DBMaxConns:              int32(dbMaxConns),
		DBMinConns:              int32(dbMinConns),
		AIEnabled:               aiEnabled,
		AIProvider:              valueOrDefault("AI_PROVIDER", "ollama"),
		AIBaseURL:               valueOrDefault("AI_BASE_URL", "http://127.0.0.1:11434"),
		AIModel:                 valueOrDefault("AI_MODEL", "qwen3:4b-instruct"),
		AIWorkerURL:             strings.TrimRight(strings.TrimSpace(os.Getenv("AI_WORKER_URL")), "/"),
		AIWorkerToken:           strings.TrimSpace(os.Getenv("AI_WORKER_TOKEN")),
		AITimeout:               time.Duration(aiTimeoutSeconds) * time.Second,
		AIMaxTokens:             aiMaxTokens,
		AIMaxConcurrent:         aiMaxConcurrent,
		ProtectionLevelsEnabled: protectionLevels,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SecureCookies && (cfg.EmailFrom == "" || cfg.ResendAPIKey == "") {
		return Config{}, errors.New("EMAIL_FROM and RESEND_API_KEY are required in production")
	}
	if cfg.SecureCookies {
		for name, value := range map[string]string{"APP_BASE_URL": cfg.AppBaseURL, "ALLOWED_ORIGIN": cfg.AllowedOrigin} {
			if parsed, err := url.Parse(value); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
				return Config{}, fmt.Errorf("%s must be an https URL in production", name)
			}
		}
	}
	cfg.Development = !cfg.SecureCookies && isLoopbackURL(cfg.AppBaseURL)
	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func isLoopbackURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
