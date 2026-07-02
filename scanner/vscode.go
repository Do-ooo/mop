package scanner

import (
	"os"
	"path/filepath"
)

type VSCodeScanner struct{}

func init() {
	Register(&VSCodeScanner{})
}

func (s *VSCodeScanner) Name() string {
	return "VS Code"
}

func (s *VSCodeScanner) Type() string {
	return "Desktop"
}

func (s *VSCodeScanner) Enabled() bool {
	return GetEnabled("VS Code")
}

func (s *VSCodeScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	desktopPath := filepath.Join(home, "Library", "Application Support", "Code")
	_, err = os.Stat(desktopPath)
	return err == nil
}

func (s *VSCodeScanner) Scan() ([]CacheItem, error) {
	var items []CacheItem
	items = append(items, scanElectronDesktop("Code")...)

	home, _ := os.UserHomeDir()
	// workspaceStorage - deep clean (workspace history)
	wsPath := filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage")
	if info, err := os.Stat(wsPath); err == nil && info.IsDir() {
		size, _ := dirSize(wsPath)
		if size > 0 {
			items = append(items, CacheItem{
				Path:        wsPath,
				Size:        size,
				Description: "Workspace storage history",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	// extensions cache
	extPath := filepath.Join(home, ".vscode", "extensions")
	if info, err := os.Stat(extPath); err == nil && info.IsDir() {
		size, _ := dirSize(extPath)
		if size > 0 {
			items = append(items, CacheItem{
				Path:        extPath,
				Size:        size,
				Description: "Installed extensions",
				ModTime:     info.ModTime(),
			})
		}
	}

	return items, nil
}
