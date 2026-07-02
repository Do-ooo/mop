package scanner

import (
	"os"
	"path/filepath"
)

type OpenCodeScanner struct{}

func init() {
	Register(&OpenCodeScanner{})
}

func (s *OpenCodeScanner) Name() string {
	return "OpenCode"
}

func (s *OpenCodeScanner) Type() string {
	return "CLI"
}

func (s *OpenCodeScanner) Enabled() bool {
	return GetEnabled("OpenCode")
}

func (s *OpenCodeScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	path := filepath.Join(home, ".local", "share", "opencode")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	macPath := filepath.Join(home, "Library", "Application Support", "opencode")
	if _, err := os.Stat(macPath); err == nil {
		return true
	}
	return false
}

func (s *OpenCodeScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var basePath string
	path1 := filepath.Join(home, ".local", "share", "opencode")
	path2 := filepath.Join(home, "Library", "Application Support", "opencode")
	if _, err := os.Stat(path1); err == nil {
		basePath = path1
	} else if _, err := os.Stat(path2); err == nil {
		basePath = path2
	} else {
		return nil, nil
	}

	var items []CacheItem

	subdirs := []struct {
		name string
		desc string
	}{
		{"log", "Logs"},
		{"snapshot", "Snapshots"},
		{"repos", "Repo clones"},
		{"tool-output", "Tool output cache"},
	}

	for _, sd := range subdirs {
		fullPath := filepath.Join(basePath, sd.name)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		size, err := dirSize(fullPath)
		if err != nil {
			continue
		}
		items = append(items, CacheItem{
			Path:        fullPath,
			Size:        size,
			Description: sd.desc,
			ModTime:     info.ModTime(),
		})
	}

	return items, nil
}
