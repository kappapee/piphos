package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// configFileName defines the relative path for the configuration file
// within the user's configuration directory.
const configFileName = "piphos/config.json"

// UserConfig represents the user-configurable settings for piphos.
// These settings are persisted to and loaded from a JSON configuration file.
type UserConfig struct {
	// Hostname is the name used to identify this machine in tender services.
	// If empty, the system hostname will be used automatically.
	Hostname string `json:"hostname"`

	// Token is the authentication token for accessing tender services.
	// For GitHub, this should be a personal access token with gist permissions.
	Token string `json:"token"`

	// Beacon specifies the preferred beacon service for IP detection.
	// If empty, a random beacon will be selected from available options.
	Beacon string `json:"beacon"`

	// Tender specifies the preferred tender service for IP storage.
	// Currently supported: "github" for GitHub Gists.
	Tender string `json:"tender"`

	// PiphosGistID stores the GitHub Gist ID used for IP storage.
	// This is automatically populated after the first successful push operation.
	PiphosGistID string `json:"piphos_gist_id"`
}

// Config holds the complete configuration for a piphos session,
// including both user settings and runtime components.
type Config struct {
	// Client is the HTTP client used for all network operations.
	// It includes appropriate timeouts and other performance settings.
	Client *http.Client

	// UserConfig contains the user-configurable settings.
	UserConfig UserConfig
}

// configLoad loads the piphos configuration from the user's configuration directory.
//
// Returns a fully initialized Config struct or an error if the configuration
// cannot be loaded or parsed.
func configLoad() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, fmt.Errorf("unable to get configuration directory: %w", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	configFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("unable to open configuration file: %w", err)
	}
	defer configFile.Close()

	config, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read from configuration file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(config, &cfg.UserConfig)
	if err != nil {
		return Config{}, fmt.Errorf("unable to get configuration options: %w", err)
	}

	if cfg.UserConfig.Hostname == "" {
		cfg.UserConfig.Hostname, err = os.Hostname()
		if err != nil {
			return Config{}, fmt.Errorf("unable to get hostname: %w", err)
		}
	}

	cfg.Client = &http.Client{Timeout: 10 * time.Second}

	return cfg, nil
}

// configSave persists the current configuration to the user's configuration directory.
//
// Returns an error if the configuration cannot be saved due to filesystem issues
// or JSON marshaling failures.
func configSave(cfg Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("unable to get configuration directory: %w", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	tempFile, err := os.CreateTemp(configDir, "piphos-config-*.tmp")
	if err != nil {
		return fmt.Errorf("unable to create temporary config file: %w", err)
	}
	tempPath := tempFile.Name()

	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	configContent, err := json.MarshalIndent(cfg.UserConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal configuration: %w", err)
	}

	if _, err := tempFile.Write(configContent); err != nil {
		return fmt.Errorf("unable to write configuration: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("unable to sync configuration: %w", err)
	}

	tempFile.Close()

	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("unable to finalize configuration: %w", err)
	}

	return nil
}
