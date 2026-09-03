package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when DATABASE_URL is empty")
	}
}

func TestLoadUsesEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_PORT", "9090")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" || cfg.AllowedOrigin != "https://example.com" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadKeepsAIDisabledByDefault(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AI_ENABLED", "false")
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_MODEL", "")
	t.Setenv("AI_TIMEOUT_SECONDS", "")
	t.Setenv("AI_MAX_OUTPUT_TOKENS", "")
	t.Setenv("AI_MAX_CONCURRENT", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AIEnabled || cfg.AIProvider != "ollama" || cfg.AIBaseURL != "http://127.0.0.1:11434" || cfg.AIModel != "qwen3:4b-instruct" {
		t.Fatalf("unexpected AI defaults: %+v", cfg)
	}
}

func TestLoadRequiresResendConfigurationInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("EMAIL_FROM", "")
	t.Setenv("RESEND_API_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected production configuration to require Resend credentials")
	}
}
