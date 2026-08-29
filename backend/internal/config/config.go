package config

import (
	"errors"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	AllowedOrigin string
}

func Load() (Config, error) {
	cfg := Config{
		Port:          valueOrDefault("API_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AllowedOrigin: valueOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
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
