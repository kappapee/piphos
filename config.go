package main

import (
	"encoding/json"
	"io"
	"log"
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
		log.Printf("unable to get configuration directory: %v\n", err)
		return Config{}, err
	}

	configPath := filepath.Join(configDir, configFileName)
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Printf("unable to open configuration file: %v\n", err)
		return Config{}, err
	}
	defer configFile.Close()

	config, err := io.ReadAll(configFile)
	if err != nil {
		log.Printf("unable to read from configuration file: %v\n", err)
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(config, &cfg.UserConfig)
	if err != nil {
		log.Printf("unable to get configuration options: %v\n", err)
		return Config{}, err
	}

	if cfg.UserConfig.Hostname == "" {
		cfg.UserConfig.Hostname, err = os.Hostname()
		if err != nil {
			log.Printf("unable to get hostname: %v\n", err)
			return Config{}, err
		}
	}

	cfg.Client = &http.Client{Timeout: 10 * time.Second}

	return cfg, nil
}

func configSave(cfg Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("unable to get configuration directory: %v\n", err)
		return err
	}

	configPath := filepath.Join(configDir, configFileName)
	configFile, err := os.Create(configPath)
	if err != nil {
		log.Printf("unable to open configuration file: %v\n", err)
		return err
	}
	defer configFile.Close()

	configContent, err := json.Marshal(cfg.UserConfig)
	if err != nil {
		log.Printf("unable to prepare configuration content: %v\n", err)
		return err
	}

	_, err = configFile.Write(configContent)
	if err != nil {
		log.Printf("unable to write to configuration file: %v\n", err)
		return err
	}
	return nil
}
