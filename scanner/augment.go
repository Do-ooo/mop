package scanner

import (
	"os"
	"path/filepath"
)

type AugmentScanner struct{}

func init() {
	Register(&AugmentScanner{})
}

func (s *AugmentScanner) Name() string {
	return "Augment"
}

func (s *AugmentScanner) Type() string {
	return "CLI"
}

func (s *AugmentScanner) Enabled() bool {
	return GetEnabled("Augment")
}

func (s *AugmentScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	augmentPath := filepath.Join(home, ".augmentcode")
	_, err = os.Stat(augmentPath)
	return err == nil
}

func (s *AugmentScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".augmentcode")

	var items []CacheItem

	// device_id - deep clean (triggers re-login)
	deviceFile := filepath.Join(basePath, "device_id")
	if info, err := os.Stat(deviceFile); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        deviceFile,
				Size:        info.Size(),
				Description: "Device ID (requires re-login)",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	// Also check device-id (alternate naming)
	deviceFile2 := filepath.Join(basePath, "device-id")
	if info, err := os.Stat(deviceFile2); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        deviceFile2,
				Size:        info.Size(),
				Description: "Device ID (requires re-login)",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	return items, nil
}
