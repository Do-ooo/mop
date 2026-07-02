package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const AutoDetectInterval = 10

var (
	configDir       string
	enabledFile     string
	enabledScanners map[string]bool
	scanCount       int
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
	Scanners  map[string]bool `json:"scanners"`
	ScanCount int             `json:"scan_count"`
}

func LoadEnabled() (map[string]bool, error) {
	config := &EnabledConfig{
		Scanners: make(map[string]bool),
	}
	data, err := os.ReadFile(enabledFile)
	if err != nil {
		if os.IsNotExist(err) {
			autoDetectFirstRun()
			SaveEnabled(enabledScanners)
			return enabledScanners, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	enabledScanners = config.Scanners
	scanCount = config.ScanCount
	return config.Scanners, nil
}

func SaveEnabled(scanners map[string]bool) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	config := &EnabledConfig{
		Scanners:  scanners,
		ScanCount: scanCount,
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

func autoDetectFirstRun() {
	for _, s := range Scanners {
		enabledScanners[s.Name()] = s.Available()
	}
	scanCount = 0
}

func IncrementScanCount() {
	scanCount++
	if scanCount >= AutoDetectInterval {
		AutoDetectNewTools()
		scanCount = 0
	}
	SaveEnabled(enabledScanners)
}

func AutoDetectNewTools() {
	changed := false
	for _, s := range Scanners {
		current, exists := enabledScanners[s.Name()]
		available := s.Available()
		if !exists {
			enabledScanners[s.Name()] = available
			if available {
				changed = true
			}
		} else if !current && available {
			enabledScanners[s.Name()] = true
			changed = true
		}
	}
	if changed {
		SaveEnabled(enabledScanners)
	}
}
