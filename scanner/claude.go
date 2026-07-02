package scanner

import (
	"os"
	"path/filepath"
)

type ClaudeCodeScanner struct{}

func init() {
	Register(&ClaudeCodeScanner{})
}

func (s *ClaudeCodeScanner) Name() string {
	return "Claude Code"
}

func (s *ClaudeCodeScanner) Type() string {
	return "CLI"
}

func (s *ClaudeCodeScanner) Enabled() bool {
	return GetEnabled("Claude Code")
}

func (s *ClaudeCodeScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".claude"))
	return err == nil
}

func (s *ClaudeCodeScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(home, ".claude")
	if _, err := os.Stat(base); err != nil {
		return nil, nil
	}
	var items []CacheItem
	cliItems := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"cache", "Cache", RiskRegular},
		{"sessions", "Sessions", RiskDeep},
		{"shell-snapshots", "Shell snapshots", RiskRegular},
		{"backups", "Backups", RiskDeep},
	}
	for _, it := range cliItems {
		fullPath := filepath.Join(base, it.name)
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
			Description: it.desc,
			ModTime:     info.ModTime(),
			Risk:        it.risk,
		})
	}
	return items, nil
}
