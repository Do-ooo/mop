package scanner

import (
	"os"
	"path/filepath"
)

type ContinueScanner struct{}

func init() {
	Register(&ContinueScanner{})
}

func (s *ContinueScanner) Name() string {
	return "Continue"
}

func (s *ContinueScanner) Type() string {
	return "CLI"
}

func (s *ContinueScanner) Enabled() bool {
	return GetEnabled("Continue")
}

func (s *ContinueScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	continuePath := filepath.Join(home, ".continue")
	_, err = os.Stat(continuePath)
	return err == nil
}

func (s *ContinueScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".continue")

	scanDirs := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"logs", "Logs", RiskRegular},
		{".utils", "Utility binaries cache", RiskRegular},
		{".diffs", "Diff files", RiskRegular},
		{"index", "Code index cache", RiskRegular},
		{"dev_data", "Dev data", RiskRegular},
		{"sessions", "Chat sessions", RiskDeep},
		{".configs", "Remote configs", RiskDeep},
		{".migrations", "Migration records", RiskRegular},
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

	return items, nil
}
