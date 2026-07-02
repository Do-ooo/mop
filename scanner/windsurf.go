package scanner

import (
	"os"
	"path/filepath"
)

type WindsurfScanner struct{}

func init() {
	Register(&WindsurfScanner{})
}

func (s *WindsurfScanner) Name() string {
	return "Windsurf"
}

func (s *WindsurfScanner) Type() string {
	return "Desktop"
}

func (s *WindsurfScanner) Enabled() bool {
	return GetEnabled("Windsurf")
}

func (s *WindsurfScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	desktopPath := filepath.Join(home, "Library", "Application Support", "Windsurf")
	_, err = os.Stat(desktopPath)
	return err == nil
}

func (s *WindsurfScanner) Scan() ([]CacheItem, error) {
	return scanElectronDesktop("Windsurf"), nil
}
