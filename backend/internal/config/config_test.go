package config_test

import (
	"testing"

	"github.com/xcreativs/gigmann/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_PORT", "8080")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.HTTPPort != 8080 {
		t.Errorf("want port 8080, got %d", cfg.HTTPPort)
	}
	if cfg.HTTPAddr() != ":8080" {
		t.Errorf("want addr :8080, got %s", cfg.HTTPAddr())
	}
	if cfg.AnthropicModel == "" {
		t.Error("expected a default Anthropic model")
	}
	if cfg.IsProduction() {
		t.Error("development should not report production")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_PORT", "not-a-number")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid HTTP_PORT")
	}
}

func TestLoadInvalidEnv(t *testing.T) {
	t.Setenv("APP_ENV", "weird")
	t.Setenv("HTTP_PORT", "8080")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid APP_ENV")
	}
}

func TestLoadProductionRequiresSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected production to require DATABASE_URL and ANTHROPIC_API_KEY")
	}
}

func TestLoadProductionValid(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("expected production")
	}
}
