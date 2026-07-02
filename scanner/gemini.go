package scanner

import (
	"os"
	"path/filepath"
)

type GeminiScanner struct{}

func init() {
	Register(&GeminiScanner{})
}

func (s *GeminiScanner) Name() string {
	return "Gemini"
}

func (s *GeminiScanner) Type() string {
	return "CLI + Desktop"
}

func (s *GeminiScanner) Enabled() bool {
	return GetEnabled("Gemini")
}

func (s *GeminiScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".gemini")
	desktopPath := filepath.Join(home, "Library", "Application Support", "Antigravity")
	_, err1 := os.Stat(cliPath)
	_, err2 := os.Stat(desktopPath)
	return err1 == nil || err2 == nil
}

func (s *GeminiScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem

	// CLI: ~/.gemini/antigravity-cli/
	cliBase := filepath.Join(home, ".gemini", "antigravity-cli")
	cliDirs := []struct {
		name string
		desc string
		risk RiskLevel
	}{
		{"cache", "Cache", RiskRegular},
		{"log", "Logs", RiskRegular},
		{"bin", "CLI binaries", RiskRegular},
		{"updater", "Updater cache", RiskRegular},
		{"conversations", "Conversation history", RiskDeep},
		{"brain", "AI memory state", RiskDeep},
		{"knowledge", "Knowledge base", RiskDeep},
		{"implicit", "Implicit sessions", RiskDeep},
	}
	for _, d := range cliDirs {
		fullPath := filepath.Join(cliBase, d.name)
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

	// CLI: history.jsonl (file, not dir)
	histFile := filepath.Join(cliBase, "history.jsonl")
	if info, err := os.Stat(histFile); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        histFile,
				Size:        info.Size(),
				Description: "History log",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	// Desktop: Antigravity (Electron app)
	items = append(items, scanElectronDesktop("Antigravity")...)

	return items, nil
}
