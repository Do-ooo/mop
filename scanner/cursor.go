package scanner

import (
	"os"
	"path/filepath"
)

type CursorScanner struct{}

func init() {
	Register(&CursorScanner{})
}

func (s *CursorScanner) Name() string {
	return "Cursor"
}

func (s *CursorScanner) Type() string {
	return "Desktop"
}

func (s *CursorScanner) Enabled() bool {
	return GetEnabled("Cursor")
}

func (s *CursorScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	desktopPath := filepath.Join(home, "Library", "Application Support", "Cursor")
	_, err = os.Stat(desktopPath)
	return err == nil
}

func (s *CursorScanner) Scan() ([]CacheItem, error) {
	return scanElectronDesktop("Cursor"), nil
}
