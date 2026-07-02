package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var (
	configDir  string
	configFile string
)

type AppConfig struct {
	TrashMode bool `json:"trashMode"`
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configDir = filepath.Join(home, ".config", "mop")
	configFile = filepath.Join(configDir, "config.json")
}

func Load() (*AppConfig, error) {
	config := &AppConfig{
		TrashMode: true,
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func Save(config *AppConfig) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}
