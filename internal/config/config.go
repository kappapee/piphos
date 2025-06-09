package config

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

type Config struct {
	Client   *http.Client
	Hostname string `json:"hostname"`
	Token    string `json:"token"`
	Beacon   string `json:"beacon"`
	Tender   string `json:"tender"`
}

func Load() (Config, error) {
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
	err = json.Unmarshal(config, &cfg)
	if err != nil {
		log.Printf("unable to get configuration options: %v\n", err)
		return Config{}, err
	}

	if cfg.Hostname == "" {
		cfg.Hostname, err = os.Hostname()
		if err != nil {
			log.Printf("unable to get hostname: %v\n", err)
			return Config{}, err
		}
	}

	cfg.Client = &http.Client{Timeout: 10 * time.Second}

	return cfg, nil
}
