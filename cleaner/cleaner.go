package cleaner

import (
	"fmt"
	"mop/scanner"
	"os"
	"path/filepath"
	"strings"
)

type CleanResult struct {
	Path    string
	Size    int64
	Success bool
	Error   string
}

type Cleaner interface {
	Clean(item scanner.CacheItem) (CleanResult, error)
}

type FileCleaner struct{}

func NewFileCleaner() *FileCleaner {
	return &FileCleaner{}
}

// validateCleanPath ensures a path is safe to delete: it must be an absolute
// location strictly inside the user's home directory, never the home directory
// itself. This guards against a misconfigured scanner handing the cleaner a
// dangerous target such as "/" or a system directory.
func validateCleanPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	abs = filepath.Clean(abs)
	home = filepath.Clean(home)
	if abs == home {
		return fmt.Errorf("refusing to clean home directory itself: %s", abs)
	}
	rel, err := filepath.Rel(home, abs)
	if err != nil {
		return fmt.Errorf("cannot resolve path relative to home: %s", abs)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("refusing to clean path outside home: %s", abs)
	}
	return nil
}
