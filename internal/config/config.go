package config

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const configFileName = "/piphos/config.json"

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
		return Config{}, err
	}

	configFile, err := os.Open(configDir + configFileName)
	if err != nil {
		return Config{}, err
	}
	defer configFile.Close()

	config, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(config, &cfg)
	if err != nil {
		return Config{}, err
	}

	if cfg.Hostname == "" {
		cfg.Hostname, err = os.Hostname()
		if err != nil {
			log.Printf("unable to set hostname: %v", err)
			return Config{}, err
		}
	}

	cfg.Client = &http.Client{Timeout: 10 * time.Second}

	return cfg, nil
}
