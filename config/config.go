package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	JWTSecret  []byte
	AdminToken string
	Port       string
	LogLevel   slog.Level
}

// Load reads required configuration from environment variables.
// Fails fast on any missing or obviously-invalid value so misconfigured
// deployments crash loudly at boot instead of limping along with defaults.
func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes (got %d)", len(secret))
	}

	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		return nil, fmt.Errorf("ADMIN_TOKEN is required")
	}
	if len(adminToken) < 16 {
		return nil, fmt.Errorf("ADMIN_TOKEN must be at least 16 bytes (got %d)", len(adminToken))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		JWTSecret:  []byte(secret),
		AdminToken: adminToken,
		Port:       port,
		LogLevel:   parseLogLevel(os.Getenv("LOG_LEVEL")),
	}, nil
}

// parseLogLevel turns a LOG_LEVEL env-var string (debug|info|warn|error,
// case-insensitive) into a slog.Level. Empty or unknown values default
// to Info so a misconfigured operator still sees something useful.
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
