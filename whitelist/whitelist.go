package whitelist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

var (
	configDir  string
	configFile string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configDir = filepath.Join(home, ".config", "mop")
	configFile = filepath.Join(configDir, "whitelist.json")
}

func Load() (map[string]bool, error) {
	wl := make(map[string]bool)
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return wl, nil
		}
		return nil, err
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return nil, err
	}
	for _, p := range paths {
		wl[p] = true
	}
	return wl, nil
}

func Save(wl map[string]bool) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	var paths []string
	for p := range wl {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	data, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

func Toggle(wl map[string]bool, path string) bool {
	if wl[path] {
		delete(wl, path)
		return false
	}
	wl[path] = true
	return true
}

func IsWhitelisted(wl map[string]bool, path string) bool {
	return wl[path]
}
