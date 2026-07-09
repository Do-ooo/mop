package scanner

import (
	"os"
	"path/filepath"
)

type MiMoCodeScanner struct{}

func init() {
	Register(&MiMoCodeScanner{})
}

func (s *MiMoCodeScanner) Name() string {
	return "MiMo Code"
}

func (s *MiMoCodeScanner) Type() string {
	return "CLI"
}

func (s *MiMoCodeScanner) Enabled() bool {
	return GetEnabled("MiMo Code")
}

func (s *MiMoCodeScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	candidates := []string{
		filepath.Join(home, ".config", "mimocode"),
		filepath.Join(home, "Library", "Application Support", "mimocode"),
		filepath.Join(home, ".local", "share", "mimocode"),
		filepath.Join(home, ".cache", "mimocode"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}

func (s *MiMoCodeScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataPaths := []string{
		filepath.Join(home, "Library", "Application Support", "mimocode"),
		filepath.Join(home, ".local", "share", "mimocode"),
	}
	cachePath := filepath.Join(home, ".cache", "mimocode")

	var items []CacheItem

	// cache dir - safe to clean
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		size, _ := dirSize(cachePath)
		if size > 0 {
			items = append(items, CacheItem{
				Path:        cachePath,
				Size:        size,
				Description: "Cache",
				ModTime:     info.ModTime(),
				Risk:        RiskRegular,
			})
		}
	}

	subdirs := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"log", "Logs", RiskRegular},
		{"snapshot", "Snapshots", RiskRegular},
		{"repos", "Repo clones", RiskRegular},
		{"tool-output", "Tool output cache", RiskRegular},
		{"sessions", "Sessions", RiskDeep},
	}
	for _, base := range dataPaths {
		for _, sd := range subdirs {
			fullPath := filepath.Join(base, sd.name)
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
				Description: sd.desc,
				ModTime:     info.ModTime(),
				Risk:        sd.risk,
			})
		}
		// auth.json - OAuth tokens (deep clean, requires re-login)
		authFile := filepath.Join(base, "auth.json")
		if info, err := os.Stat(authFile); err == nil && !info.IsDir() {
			if info.Size() > 0 {
				items = append(items, CacheItem{
					Path:        authFile,
					Size:        info.Size(),
					Description: "Auth tokens (requires re-login)",
					ModTime:     info.ModTime(),
					Risk:        RiskDeep,
				})
			}
		}
	}

	return items, nil
}
