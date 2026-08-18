// Package cmd implements the custodian-cli command tree.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	flagURL   string
	flagToken string
)

// rootCmd is the base command; every subcommand inherits --url and --token.
var rootCmd = &cobra.Command{
	Use:   "custodian",
	Short: "Custodian — self-hosted PaaS with full ownership",
	Long: `Custodian is a self-hosted platform-as-a-service that brings the
Render/Heroku developer experience to your own infrastructure.

Deploy apps, manage services, stream logs and configure autoscaling —
all from a single control plane.`,
	SilenceUsage: true,
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "Custodian control plane URL (default: from config or http://localhost:8080)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (default: from config)")
}

func init() {
	// Config is loaded lazily so `custodian login` works before any config exists.
}

// Client bundles the control plane URL and credential for API calls.
type Client struct {
	URL   string
	Token string
}

// NewClient resolves connection settings from flags, then the local config.
func NewClient() (*Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	url := flagURL
	if url == "" {
		url = cfg.URL
	}
	if url == "" {
		url = "http://localhost:8080"
	}
	token := flagToken
	if token == "" {
		token = cfg.Token
	}
	if token == "" {
		return nil, fmt.Errorf("not authenticated; run `custodian login` or pass --token")
	}
	return &Client{URL: url, Token: token}, nil
}

type fileConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".custodian", "config.json"), nil
}

func loadConfig() (*fileConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &fileConfig{}, nil
	}
	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

func saveConfig(cfg *fileConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
