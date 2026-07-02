package scanner

import (
	"os"
	"path/filepath"
)

type CodeBuddyScanner struct{}

func init() {
	Register(&CodeBuddyScanner{})
}

func (s *CodeBuddyScanner) Name() string {
	return "CodeBuddy"
}

func (s *CodeBuddyScanner) Type() string {
	return "CLI + Desktop"
}

func (s *CodeBuddyScanner) Enabled() bool {
	return GetEnabled("CodeBuddy")
}

func (s *CodeBuddyScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".codebuddy")
	desktopPath := filepath.Join(home, "Library", "Application Support", "CodeBuddy")
	_, err1 := os.Stat(cliPath)
	_, err2 := os.Stat(desktopPath)
	return err1 == nil || err2 == nil
}

func (s *CodeBuddyScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem
	cliPath := filepath.Join(home, ".codebuddy")
	if _, err := os.Stat(cliPath); err == nil {
		cliItems := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"logs", "Logs", RiskRegular},
		{"sessions", "Sessions", RiskDeep},
		{"shell-snapshots", "Shell snapshots", RiskRegular},
		{"traces", "Traces", RiskRegular},
	}
	for _, it := range cliItems {
		fullPath := filepath.Join(cliPath, it.name)
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
	}
	items = append(items, scanElectronDesktop("CodeBuddy")...)
	return items, nil
}
