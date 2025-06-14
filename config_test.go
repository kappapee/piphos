package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "piphos-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	piphosDir := filepath.Join(tempDir, "piphos")
	if err := os.MkdirAll(piphosDir, 0755); err != nil {
		t.Fatalf("Failed to create piphos directory: %v", err)
	}

	originalUserConfigDir := userConfigDirFunc
	userConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		userConfigDirFunc = originalUserConfigDir
	}()

	testConfig := Config{
		UserConfig: UserConfig{
			Hostname:     "test-host",
			Token:        "test-token",
			Beacon:       "test-beacon",
			Tender:       "github",
			PiphosGistID: "test-gist-id",
		},
	}

	t.Run("SaveConfig", func(t *testing.T) {
		if err := configSave(testConfig); err != nil {
			t.Errorf("configSave failed: %v", err)
		}

		configPath := filepath.Join(tempDir, configFileName)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Errorf("Failed to read config file: %v", err)
		}

		var savedConfig UserConfig
		if err := json.Unmarshal(content, &savedConfig); err != nil {
			t.Errorf("Failed to parse saved config: %v", err)
		}

		if savedConfig.Hostname != testConfig.UserConfig.Hostname {
			t.Errorf("Expected hostname %s, got %s", testConfig.UserConfig.Hostname, savedConfig.Hostname)
		}
		if savedConfig.Token != testConfig.UserConfig.Token {
			t.Errorf("Expected token %s, got %s", testConfig.UserConfig.Token, savedConfig.Token)
		}
	})

	t.Run("LoadConfig", func(t *testing.T) {
		loadedConfig, err := configLoad()
		if err != nil {
			t.Errorf("configLoad failed: %v", err)
		}

		if loadedConfig.UserConfig.Hostname != testConfig.UserConfig.Hostname {
			t.Errorf("Expected hostname %s, got %s", testConfig.UserConfig.Hostname, loadedConfig.UserConfig.Hostname)
		}
		if loadedConfig.UserConfig.Token != testConfig.UserConfig.Token {
			t.Errorf("Expected token %s, got %s", testConfig.UserConfig.Token, loadedConfig.UserConfig.Token)
		}
		if loadedConfig.Client == nil {
			t.Error("HTTP client was not initialized")
		}
	})

	t.Run("LoadConfigWithEmptyHostname", func(t *testing.T) {
		emptyHostnameConfig := Config{
			UserConfig: UserConfig{
				Token:  "test-token",
				Beacon: "test-beacon",
				Tender: "github",
			},
		}

		if err := configSave(emptyHostnameConfig); err != nil {
			t.Fatalf("Failed to save config with empty hostname: %v", err)
		}

		loadedConfig, err := configLoad()
		if err != nil {
			t.Fatalf("Failed to load config with empty hostname: %v", err)
		}

		if loadedConfig.UserConfig.Hostname == "" {
			t.Error("Expected system hostname to be set, got empty string")
		}
	})
}

func TestConfigLoadErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "piphos-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalUserConfigDir := userConfigDirFunc
	userConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		userConfigDirFunc = originalUserConfigDir
	}()

	t.Run("LoadNonExistentConfig", func(t *testing.T) {
		_, err := configLoad()
		if err == nil {
			t.Error("Expected error when loading non-existent config, got nil")
		}
	})

	t.Run("LoadMalformedJSON", func(t *testing.T) {
		configPath := filepath.Join(tempDir, configFileName)
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.WriteFile(configPath, []byte("{malformed json}"), 0644); err != nil {
			t.Fatalf("Failed to write malformed config: %v", err)
		}

		_, err := configLoad()
		if err == nil {
			t.Error("Expected error when loading malformed JSON, got nil")
		}
	})
}
