package scanner

import (
	"os"
	"path/filepath"
)

type CopilotScanner struct{}

func init() {
	Register(&CopilotScanner{})
}

func (s *CopilotScanner) Name() string {
	return "Copilot"
}

func (s *CopilotScanner) Type() string {
	return "CLI"
}

func (s *CopilotScanner) Enabled() bool {
	return GetEnabled("Copilot")
}

func (s *CopilotScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".config", "github-copilot")
	_, err = os.Stat(cliPath)
	return err == nil
}

func (s *CopilotScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, ".config", "github-copilot")

	var items []CacheItem

	// hosts.json - OAuth token (deep clean)
	hostsFile := filepath.Join(basePath, "hosts.json")
	if info, err := os.Stat(hostsFile); err == nil && !info.IsDir() {
		if info.Size() > 0 {
			items = append(items, CacheItem{
				Path:        hostsFile,
				Size:        info.Size(),
				Description: "OAuth token (requires re-login)",
				ModTime:     info.ModTime(),
				Risk:        RiskDeep,
			})
		}
	}

	return items, nil
}
