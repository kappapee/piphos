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

const configFileName = "piphos/config.json"

type UserConfig struct {
	Hostname     string `json:"hostname"`
	Token        string `json:"token"`
	Beacon       string `json:"beacon"`
	Tender       string `json:"tender"`
	PiphosGistID string `json:"piphos_gist_id"`
}

type Config struct {
	Client     *http.Client
	UserConfig UserConfig
}

func configLoad() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, fmt.Errorf("unable to get configuration directory: %w\n", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	configFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("unable to open configuration file: %w\n", err)
	}
	defer configFile.Close()

	config, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read from configuration file: %w\n", err)
	}

	var cfg Config
	err = json.Unmarshal(config, &cfg.UserConfig)
	if err != nil {
		return Config{}, fmt.Errorf("unable to get configuration options: %w\n", err)
	}

	if cfg.UserConfig.Hostname == "" {
		cfg.UserConfig.Hostname, err = os.Hostname()
		if err != nil {
			return Config{}, fmt.Errorf("unable to get hostname: %w\n", err)
		}
	}

	cfg.Client = &http.Client{Timeout: 10 * time.Second}

	return cfg, nil
}

func configSave(cfg Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("unable to get configuration directory: %w\n", err)
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
