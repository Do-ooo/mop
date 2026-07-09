package scanner

import (
	"os"
	"path/filepath"
)

type TraeScanner struct{}

func init() {
	Register(&TraeScanner{})
}

func (s *TraeScanner) Name() string {
	return "Trae"
}

func (s *TraeScanner) Type() string {
	return "CLI + Desktop"
}

func (s *TraeScanner) Enabled() bool {
	return GetEnabled("Trae")
}

func (s *TraeScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".trae")
	desktopPath1 := filepath.Join(home, "Library", "Application Support", "TRAE SOLO CN")
	desktopPath2 := filepath.Join(home, "Library", "Application Support", "Trae")
	_, err1 := os.Stat(cliPath)
	_, err2 := os.Stat(desktopPath1)
	_, err3 := os.Stat(desktopPath2)
	return err1 == nil || err2 == nil || err3 == nil
}

func (s *TraeScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem

	cliPath := filepath.Join(home, ".trae")
	if _, err := os.Stat(cliPath); err == nil {
		cliItems := []struct {
			name string
			desc string
			risk RiskLevel
		}{
			{"logs", "Logs", RiskRegular},
			{"cache", "Cache", RiskRegular},
			{"tmp", "Temp", RiskRegular},
			{"shell-snapshots", "Shell snapshots", RiskRegular},
			{"sessions", "Sessions", RiskDeep},
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

	desktopNames := []string{"TRAE SOLO CN", "Trae"}
	for _, name := range desktopNames {
		if _, err := os.Stat(filepath.Join(home, "Library", "Application Support", name)); err == nil {
			items = append(items, scanElectronDesktop(name)...)
			break
		}
	}
	return items, nil
}
