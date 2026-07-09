package scanner

import (
	"os"
	"path/filepath"
)

type CodexScanner struct{}

func init() {
	Register(&CodexScanner{})
}

func (s *CodexScanner) Name() string {
	return "Codex"
}

func (s *CodexScanner) Type() string {
	return "CLI + Desktop"
}

func (s *CodexScanner) Enabled() bool {
	return GetEnabled("Codex")
}

func (s *CodexScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".codex")
	desktopPath := filepath.Join(home, "Library", "Application Support", "Codex")
	_, err1 := os.Stat(cliPath)
	_, err2 := os.Stat(desktopPath)
	return err1 == nil || err2 == nil
}

func (s *CodexScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var items []CacheItem

	cliPath := filepath.Join(home, ".codex")
	if _, err := os.Stat(cliPath); err == nil {
		cliItems := []struct {
			name string
			desc string
			risk RiskLevel
		}{
			{".tmp", "Temp files", RiskRegular},
			{"cache", "Cache", RiskRegular},
			{"tmp", "Temp", RiskRegular},
			{"sessions", "Sessions", RiskDeep},
			{"archived_sessions", "Archived sessions", RiskDeep},
			{"shell_snapshots", "Shell snapshots", RiskRegular},
			{"log", "Logs", RiskRegular},
		}
		for _, it := range cliItems {
			fullPath := filepath.Join(cliPath, it.name)
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			size, err := dirSize(fullPath)
			if err != nil {
				continue
			}
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

	desktopPath := filepath.Join(home, "Library", "Application Support", "Codex")
	if _, err := os.Stat(desktopPath); err == nil {
		desktopItems := []struct {
			name string
			desc string
		}{
			{"Cache", "Browser cache"},
			{"GPUCache", "GPU cache"},
			{"Code Cache", "Code cache"},
		}
		for _, it := range desktopItems {
			fullPath := filepath.Join(desktopPath, it.name)
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			size, err := dirSize(fullPath)
			if err != nil {
				continue
			}
			if size == 0 {
				continue
			}
			items = append(items, CacheItem{
				Path:        fullPath,
				Size:        size,
				Description: it.desc + " (desktop)",
				ModTime:     info.ModTime(),
			})
		}
	}

	return items, nil
}
