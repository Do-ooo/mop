package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type JetBrainsScanner struct{}

func init() {
	Register(&JetBrainsScanner{})
}

func (s *JetBrainsScanner) Name() string {
	return "JetBrains"
}

func (s *JetBrainsScanner) Type() string {
	return "Desktop"
}

func (s *JetBrainsScanner) Enabled() bool {
	return GetEnabled("JetBrains")
}

func (s *JetBrainsScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cachesPath := filepath.Join(home, "Library", "Caches", "JetBrains")
	supportPath := filepath.Join(home, "Library", "Application Support", "JetBrains")
	_, err1 := os.Stat(cachesPath)
	_, err2 := os.Stat(supportPath)
	return err1 == nil || err2 == nil
}

func (s *JetBrainsScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem

	// ~/Library/Caches/JetBrains/<product><version>/
	cachesBase := filepath.Join(home, "Library", "Caches", "JetBrains")
	if entries, err := os.ReadDir(cachesBase); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			productDir := filepath.Join(cachesBase, entry.Name())
			size, _ := dirSize(productDir)
			if size == 0 {
				continue
			}
			info, _ := os.Stat(productDir)
			items = append(items, CacheItem{
				Path:        productDir,
				Size:        size,
				Description: "Cache (" + entry.Name() + ")",
				ModTime:     info.ModTime(),
				Risk:        RiskRegular,
			})
		}
	}

	// ~/Library/Logs/JetBrains/<product><version>/
	logsBase := filepath.Join(home, "Library", "Logs", "JetBrains")
	if entries, err := os.ReadDir(logsBase); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			productDir := filepath.Join(logsBase, entry.Name())
			size, _ := dirSize(productDir)
			if size == 0 {
				continue
			}
			info, _ := os.Stat(productDir)
			items = append(items, CacheItem{
				Path:        productDir,
				Size:        size,
				Description: "Logs (" + entry.Name() + ")",
				ModTime:     info.ModTime(),
				Risk:        RiskRegular,
			})
		}
	}

	// Deep clean: ~/Library/Application Support/JetBrains/<product><version>/options/
	supportBase := filepath.Join(home, "Library", "Application Support", "JetBrains")
	if entries, err := os.ReadDir(supportBase); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			optionsDir := filepath.Join(supportBase, entry.Name(), "options")
			if info, err := os.Stat(optionsDir); err == nil && info.IsDir() {
				size, _ := dirSize(optionsDir)
				if size > 0 {
					productName := strings.TrimSuffix(entry.Name(), "/options")
					items = append(items, CacheItem{
						Path:        optionsDir,
						Size:        size,
						Description: "Options (" + productName + ")",
						ModTime:     info.ModTime(),
						Risk:        RiskDeep,
					})
				}
			}
		}
	}

	return items, nil
}
