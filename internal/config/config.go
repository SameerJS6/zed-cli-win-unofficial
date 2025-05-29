package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ZedPath            string `json:"zedPath"`
	ContextMenuEnabled bool   `json:"contextMenuEnabled"`
}

// ConfigPath returns the path of the configuration file.
func ConfigPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "zed-cli-win-unofficial", "config.json")
}

// SaveConfig saves the configuration to disk (config.json)
func SaveConfig(config *Config) error {
	configPath := ConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create required directories: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create the config file: %w", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode the config file: %w", err)
	}

	fmt.Printf("✅ Config saved successfully at: %s\n", configPath)
	return nil
}

// LoadConfig loads the configuration file
func LoadConfig() (*Config, error) {
	configPath := ConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to locate the configure file: %w", err)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the configure file: %w", err)
	}

	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to decode the configure file: %w", err)
	}
	return &config, nil
}
