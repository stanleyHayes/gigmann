// Package config loads and validates application configuration from the
// environment (12-factor). It fails fast on invalid configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Application environments.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Config holds validated runtime configuration.
type Config struct {
	AppEnv          string
	HTTPPort        int
	LogLevel        string
	DatabaseURL     string
	RedisURL        string
	AnthropicAPIKey string
	AnthropicModel  string
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
	}

	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid HTTP_PORT: %w", err)
	}
	cfg.HTTPPort = port

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
	if c.AppEnv != EnvDevelopment {
		if c.DatabaseURL == "" {
			return errors.New("config: DATABASE_URL is required outside development")
		}
		if c.AnthropicAPIKey == "" {
			return errors.New("config: ANTHROPIC_API_KEY is required outside development")
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
