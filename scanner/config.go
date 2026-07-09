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

// AutoDetectNewTools reconciles the enabled set with what's actually installed:
// it checks tools that are newly or re-available and unchecks tools that are no
// longer installed, so dead entries don't linger and newly added tools show up.
func AutoDetectNewTools() {
	changed := false
	for _, s := range Scanners {
		name := s.Name()
		current, exists := enabledScanners[name]
		available := s.Available()
		switch {
		case !exists:
			// Scanner not seen before (e.g. added in an update): adopt its availability.
			enabledScanners[name] = available
			if available {
				changed = true
			}
		case current && !available:
			// Was enabled but the tool is gone: uncheck it.
			enabledScanners[name] = false
			changed = true
		case !current && available:
			// Was unchecked but now installed (reinstalled or new): check it.
			enabledScanners[name] = true
			changed = true
		}
	}
	if changed {
		SaveEnabled(enabledScanners)
	}
}
