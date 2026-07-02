package scanner

import (
	"os"
	"path/filepath"
)

type QoderScanner struct{}

func init() {
	Register(&QoderScanner{})
}

func (s *QoderScanner) Name() string {
	return "Qoder"
}

func (s *QoderScanner) Type() string {
	return "CLI + Desktop"
}

func (s *QoderScanner) Enabled() bool {
	return GetEnabled("Qoder")
}

func (s *QoderScanner) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cliPath := filepath.Join(home, ".qoder")
	desktopPath1 := filepath.Join(home, "Library", "Application Support", "Qoder CN")
	desktopPath2 := filepath.Join(home, "Library", "Application Support", "Qoder")
	_, err1 := os.Stat(cliPath)
	_, err2 := os.Stat(desktopPath1)
	_, err3 := os.Stat(desktopPath2)
	return err1 == nil || err2 == nil || err3 == nil
}

func (s *QoderScanner) Scan() ([]CacheItem, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var items []CacheItem
	cliPath := filepath.Join(home, ".qoder")
	if _, err := os.Stat(cliPath); err == nil {
		cliItems := []struct {
			name string
			desc string
		}{
			{"cache", "Cache"},
		}
		for _, it := range cliItems {
			fullPath := filepath.Join(cliPath, it.name)
			info, err := os.Stat(fullPath)
			if err != nil || !info.IsDir() {
				continue
			}
			size, _ := dirSize(fullPath)
			if size == 0 {
				continue
			}
			items = append(items, CacheItem{
				Path:        fullPath,
				Size:        size,
				Description: it.desc,
				ModTime:     info.ModTime(),
			})
		}
	}
	desktopNames := []string{"Qoder CN", "Qoder"}
	for _, name := range desktopNames {
		if _, err := os.Stat(filepath.Join(home, "Library", "Application Support", name)); err == nil {
			items = append(items, scanElectronDesktop(name)...)
			break
		}
	}
	return items, nil
}
