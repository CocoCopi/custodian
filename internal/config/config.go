// Package config loads and validates Custodian control plane configuration
// from environment variables (12-factor style).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the control plane.
type Config struct {
	// HTTP server
	Addr            string
	PublicURL       string
	ShutdownTimeout time.Duration

	// Database (PostgreSQL)
	DatabaseURL string

	// Cache / queue (Redis)
	RedisAddr     string
	RedisPassword string

	// Authentication
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	JWTSecret        string
	TokenTTL         time.Duration

	// Deployment engine
	Engine string // "compose" (default) or "k3s"

	// Runtime paths
	DeployRoot string
	DataDir    string

	// Observability
	OTLPEndpoint string
}

// Load reads configuration from the environment. Required keys must be
// present, otherwise an error is returned so the process fails fast.
func Load() (*Config, error) {
	cfg := &Config{
		Addr:             getEnv("CUSTODIAN_ADDR", ":8080"),
		PublicURL:        strings.TrimSuffix(getEnv("CUSTODIAN_PUBLIC_URL", "http://localhost:8080"), "/"),
		ShutdownTimeout:  getDuration("CUSTODIAN_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:      getEnv("CUSTODIAN_DATABASE_URL", "postgres://custodian:custodian@localhost:5432/custodian?sslmode=disable"),
		RedisAddr:        getEnv("CUSTODIAN_REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("CUSTODIAN_REDIS_PASSWORD"),
		OIDCIssuer:       os.Getenv("CUSTODIAN_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("CUSTODIAN_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("CUSTODIAN_OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:  getEnv("CUSTODIAN_OIDC_REDIRECT_URL", ""),
		JWTSecret:        os.Getenv("CUSTODIAN_JWT_SECRET"),
		TokenTTL:         getDuration("CUSTODIAN_TOKEN_TTL", 24*time.Hour),
		Engine:           getEnv("CUSTODIAN_ENGINE", "compose"),
		DeployRoot:       getEnv("CUSTODIAN_DEPLOY_ROOT", "./data/deploy"),
		DataDir:          getEnv("CUSTODIAN_DATA_DIR", "./data"),
		OTLPEndpoint:     os.Getenv("CUSTODIAN_OTLP_ENDPOINT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("CUSTODIAN_DATABASE_URL is required")
	}
	if cfg.Engine != "compose" && cfg.Engine != "k3s" {
		return nil, fmt.Errorf("CUSTODIAN_ENGINE must be 'compose' or 'k3s', got %q", cfg.Engine)
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

// GetInt is exported for tests and internal helpers.
func GetInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
