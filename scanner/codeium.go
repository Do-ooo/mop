package scanner

import (
	"os"
	"path/filepath"
)

type CodeiumScanner struct{}

func init() {
	Register(&CodeiumScanner{})
}

func (s *CodeiumScanner) Name() string {
	return "Codeium"
}

func (s *CodeiumScanner) Type() string {
	return "CLI"
}

func (s *CodeiumScanner) Enabled() bool {
	return GetEnabled("Codeium")
}

func (s *CodeiumScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	codeiumPath := filepath.Join(home, ".codeium")
	_, err = os.Stat(codeiumPath)
	return err == nil
}

func (s *CodeiumScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".codeium")

	var items []CacheItem

	// Scan versioned subdirectories (language server binaries)
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return items, nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(basePath, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		size, _ := dirSize(fullPath)
		if size == 0 {
			continue
		}
		items = append(items, CacheItem{
			Path:        fullPath,
			Size:        size,
			Description: "Language server binary (v" + entry.Name() + ")",
			ModTime:     info.ModTime(),
			Risk:        RiskRegular,
		})
	}

	return items, nil
}
