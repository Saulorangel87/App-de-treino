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

func TestLoadRejectsUnknownAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DATABASE_URL", "postgres://example")
	if _, err := Load(); err == nil {
		t.Fatal("expected an unknown APP_ENV to be rejected")
	}
}

func TestLoadRequiresHTTPSURLsInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("EMAIL_FROM", "Cadencia <no-reply@example.com>")
	t.Setenv("RESEND_API_KEY", "key")
	t.Setenv("APP_BASE_URL", "http://example.com")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	if _, err := Load(); err == nil {
		t.Fatal("expected production to require an https APP_BASE_URL")
	}
	t.Setenv("APP_BASE_URL", "https://example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Development || !cfg.SecureCookies {
		t.Fatalf("production must never enable development links: %+v", cfg)
	}
}

func TestDevelopmentLinksRequireLoopbackURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_BASE_URL", "https://cadencia.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Development {
		t.Fatal("a public APP_BASE_URL must not enable development links even when APP_ENV is unset")
	}
	t.Setenv("APP_BASE_URL", "http://localhost:3000")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Development {
		t.Fatal("expected development links for a loopback base URL outside production")
	}
}

func TestLoadValidatesDatabasePool(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DB_MAX_CONNS", "4")
	t.Setenv("DB_MIN_CONNS", "5")
	if _, err := Load(); err == nil {
		t.Fatal("expected DB_MIN_CONNS above DB_MAX_CONNS to be rejected")
	}
}
