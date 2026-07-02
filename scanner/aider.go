package scanner

import (
	"os"
	"path/filepath"
)

type AiderScanner struct{}

func init() {
	Register(&AiderScanner{})
}

func (s *AiderScanner) Name() string {
	return "Aider"
}

func (s *AiderScanner) Type() string {
	return "CLI"
}

func (s *AiderScanner) Enabled() bool {
	return GetEnabled("Aider")
}

func (s *AiderScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	aiderPath := filepath.Join(home, ".aider")
	_, err = os.Stat(aiderPath)
	return err == nil
}

func (s *AiderScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".aider")

	scanDirs := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"oauth-keys.env", "OAuth keys", RiskDeep},
	}

	var items []CacheItem
	for _, d := range scanDirs {
		fullPath := filepath.Join(basePath, d.name)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
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
		} else {
			if info.Size() > 0 {
				items = append(items, CacheItem{
					Path:        fullPath,
					Size:        info.Size(),
					Description: d.desc,
					ModTime:     info.ModTime(),
					Risk:        d.risk,
				})
			}
		}
	}

	return items, nil
}
