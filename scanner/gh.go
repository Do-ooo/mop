package scanner

import (
	"os"
	"path/filepath"
)

type GitHubCLIScanner struct{}

func init() {
	Register(&GitHubCLIScanner{})
}

func (s *GitHubCLIScanner) Name() string {
	return "GitHub CLI"
}

func (s *GitHubCLIScanner) Type() string {
	return "CLI"
}

func (s *GitHubCLIScanner) Enabled() bool {
	return GetEnabled("GitHub CLI")
}

func (s *GitHubCLIScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	ghPath := filepath.Join(home, ".config", "gh")
	_, err = os.Stat(ghPath)
	return err == nil
}

func (s *GitHubCLIScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem

	// ~/.cache/gh/ - regular cache
	cachePath := filepath.Join(home, ".cache", "gh")
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		size, _ := dirSize(cachePath)
		if size > 0 {
			items = append(items, CacheItem{
				Path:        cachePath,
				Size:        size,
				Description: "CLI cache",
				ModTime:     info.ModTime(),
				Risk:        RiskRegular,
			})
		}
	}

	// ~/.local/state/gh/device-id - deep clean (triggers re-auth)
	deviceFile := filepath.Join(home, ".local", "state", "gh", "device-id")
	if info, err := os.Stat(deviceFile); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        deviceFile,
				Size:        info.Size(),
				Description: "Device ID (requires re-auth)",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	// ~/.config/gh/hosts.yml - deep clean (OAuth token, requires re-login)
	hostsFile := filepath.Join(home, ".config", "gh", "hosts.yml")
	if info, err := os.Stat(hostsFile); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        hostsFile,
				Size:        info.Size(),
				Description: "Auth tokens (requires re-login)",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	return items, nil
}
