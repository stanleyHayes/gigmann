// Package config loads and validates application configuration from the
// environment (12-factor). It fails fast on invalid configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Application environments.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// devJWTSecret is a non-secret placeholder used only in development.
const devJWTSecret = "dev-insecure-change-me" //nolint:gosec // non-secret development placeholder

// Config holds validated runtime configuration.
type Config struct {
	AppEnv             string
	HTTPPort           int
	LogLevel           string
	DatabaseURL        string
	RedisURL           string
	AnthropicAPIKey    string
	AnthropicModel     string
	VoyageAPIKey       string
	VoyageModel        string
	SentryDSN          string
	JWTSecret          string
	CORSAllowedOrigins []string
	TrustProxy         bool   // trust X-Forwarded-For for the client IP (set only behind a trusted proxy, e.g. Render)
	VAPIDPublicKey     string // Web Push (GEC-69); push is disabled when unset
	VAPIDPrivateKey    string
	VAPIDSubject       string
	Flags              Flags
}

// Flags are runtime feature toggles (12-factor: set via FEATURE_* env vars).
// They default ON so the full product runs unless a feature is explicitly disabled.
type Flags struct {
	AINarration    bool // use Claude for the brief/Ask when a key is set (else local narrator)
	FacilitySearch bool // embed facilities + serve NL facility search
}

// Load reads configuration from the environment and validates it.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:          getEnv("APP_ENV", EnvDevelopment),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:  getEnv("ANTHROPIC_MODEL", "claude-sonnet-4-6"),
		VoyageAPIKey:    os.Getenv("VOYAGE_API_KEY"),
		VoyageModel:     getEnv("VOYAGE_MODEL", "voyage-3.5-lite"),
		SentryDSN:       os.Getenv("SENTRY_DSN"),
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:    os.Getenv("VAPID_SUBJECT"),
	}

	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid HTTP_PORT: %w", err)
	}
	cfg.HTTPPort = port

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	cfg.CORSAllowedOrigins = splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))
	cfg.TrustProxy = getBool("TRUST_PROXY", false)
	cfg.Flags = Flags{
		AINarration:    getBool("FEATURE_AI_NARRATION", true),
		FacilitySearch: getBool("FEATURE_FACILITY_SEARCH", true),
	}
	if cfg.AppEnv == EnvDevelopment && cfg.JWTSecret == "" {
		cfg.JWTSecret = devJWTSecret
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// HTTPAddr returns the server listen address.
func (c Config) HTTPAddr() string { return fmt.Sprintf(":%d", c.HTTPPort) }

// IsProduction reports whether the app runs in production.
func (c Config) IsProduction() bool { return c.AppEnv == EnvProduction }

func (c Config) validate() error {
	switch c.AppEnv {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("config: invalid APP_ENV %q", c.AppEnv)
	}
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		return fmt.Errorf("config: HTTP_PORT out of range: %d", c.HTTPPort)
	}
	// JWT_SECRET is the only hard requirement outside development — there is no
	// safe default for token signing. DATABASE_URL and ANTHROPIC_API_KEY are
	// optional: the app falls back to in-memory repositories and the
	// deterministic local narrator when they are absent.
	if c.AppEnv != EnvDevelopment && c.JWTSecret == "" {
		return errors.New("config: JWT_SECRET is required outside development")
	}
	// Defence in depth: never let the well-known development placeholder sign
	// tokens in staging/production even if it is set explicitly.
	if c.AppEnv != EnvDevelopment && c.JWTSecret == devJWTSecret {
		return errors.New("config: JWT_SECRET must not be the development placeholder outside development")
	}
	return nil
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getBool parses a boolean feature flag; unset/blank/unparseable falls back.
func getBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return b
}
