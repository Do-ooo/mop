package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var (
	configDir     string
	enabledFile   string
	enabledScanners map[string]bool
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configDir = filepath.Join(home, ".config", "mop")
	enabledFile = filepath.Join(configDir, "enabled_scanners.json")
	enabledScanners = make(map[string]bool)
}

type EnabledConfig struct {
	Scanners map[string]bool `json:"scanners"`
}

func LoadEnabled() (map[string]bool, error) {
	config := &EnabledConfig{
		Scanners: make(map[string]bool),
	}
	data, err := os.ReadFile(enabledFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config.Scanners, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	enabledScanners = config.Scanners
	return config.Scanners, nil
}

func SaveEnabled(scanners map[string]bool) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	config := &EnabledConfig{
		Scanners: scanners,
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(enabledFile, data, 0644)
}

func GetEnabled(name string) bool {
	if len(enabledScanners) == 0 {
		return true
	}
	if enabled, ok := enabledScanners[name]; ok {
		return enabled
	}
	return false
}

func SetEnabled(name string, enabled bool) {
	enabledScanners[name] = enabled
}

func GetAllEnabled() map[string]bool {
	return enabledScanners
}

func SetEnabledFromMap(m map[string]bool) {
	enabledScanners = m
}
