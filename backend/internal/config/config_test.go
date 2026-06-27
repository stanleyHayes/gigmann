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

func TestLoadProductionRequiresJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://x") // present — proves DB is not the requirement
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("JWT_SECRET", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected production to require JWT_SECRET")
	}
}

func TestLoadProductionMinimal(t *testing.T) {
	// JWT_SECRET alone suffices outside dev; DATABASE_URL and ANTHROPIC_API_KEY
	// are optional (in-memory + local-narrator fallbacks).
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("JWT_SECRET", "prod-signing-secret")
	if _, err := config.Load(); err != nil {
		t.Fatalf("expected JWT_SECRET to suffice in production: %v", err)
	}
}

func TestLoadProductionValid(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("JWT_SECRET", "prod-signing-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("expected production")
	}
}

func TestFlagsDefaultOn(t *testing.T) {
	t.Setenv("JWT_SECRET", "x")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.Flags.AINarration || !cfg.Flags.FacilitySearch {
		t.Errorf("flags should default on, got %+v", cfg.Flags)
	}
}

func TestFlagsDisable(t *testing.T) {
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("FEATURE_AI_NARRATION", "false")
	t.Setenv("FEATURE_FACILITY_SEARCH", "0")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Flags.AINarration {
		t.Error("FEATURE_AI_NARRATION=false should disable AINarration")
	}
	if cfg.Flags.FacilitySearch {
		t.Error("FEATURE_FACILITY_SEARCH=0 should disable FacilitySearch")
	}
}
