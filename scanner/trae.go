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
	desktopNames := []string{"TRAE SOLO CN", "Trae"}
	for _, name := range desktopNames {
		if _, err := os.Stat(filepath.Join(home, "Library", "Application Support", name)); err == nil {
			items = append(items, scanElectronDesktop(name)...)
			break
		}
	}
	return items, nil
}
