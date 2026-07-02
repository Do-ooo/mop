package scanner

import (
	"os"
	"path/filepath"
)

type SupermavenScanner struct{}

func init() {
	Register(&SupermavenScanner{})
}

func (s *SupermavenScanner) Name() string {
	return "Supermaven"
}

func (s *SupermavenScanner) Type() string {
	return "CLI"
}

func (s *SupermavenScanner) Enabled() bool {
	return GetEnabled("Supermaven")
}

func (s *SupermavenScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	smPath := filepath.Join(home, ".supermaven")
	_, err = os.Stat(smPath)
	return err == nil
}

func (s *SupermavenScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".supermaven")

	scanDirs := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"logs", "Logs", RiskRegular},
		{"cache", "Cache", RiskRegular},
	}

	var items []CacheItem
	for _, d := range scanDirs {
		fullPath := filepath.Join(basePath, d.name)
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}
		size, _ := dirSize(fullPath)
		if size == 0 {
			continue
		}
		items = append(items, CacheItem{
			Path:        fullPath,
			Size:        size,
			Description: d.desc,
			ModTime:     info.ModTime(),
			Risk:        d.risk,
		})
	}

	// device-id - deep clean (triggers re-login)
	deviceFile := filepath.Join(basePath, "device-id")
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

	return items, nil
}
