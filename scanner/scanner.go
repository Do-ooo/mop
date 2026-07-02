package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

type RiskLevel int

const (
	RiskRegular RiskLevel = iota
	RiskDeep
)

type CacheItem struct {
	Path        string
	Size        int64
	Description string
	ModTime     time.Time
	Risk        RiskLevel
}

type ToolGroup struct {
	Name     string
	Type     string
	Items    []CacheItem
	TotalSize int64
}

type ToolScanner interface {
	Name() string
	Type() string
	Enabled() bool
	Available() bool
	Scan() ([]CacheItem, error)
}

var Scanners []ToolScanner

func Register(s ToolScanner) {
	Scanners = append(Scanners, s)
}

func ScanAll() ([]ToolGroup, error) {
	var groups []ToolGroup
	for _, s := range Scanners {
		if !s.Enabled() {
			continue
		}
		if !s.Available() {
			continue
		}
		items, err := s.Scan()
		if err != nil {
			continue
		}
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].Size > items[j].Size
		})
		var total int64
		for _, item := range items {
			total += item.Size
		}
		groups = append(groups, ToolGroup{
			Name:      s.Name(),
			Type:      s.Type(),
			Items:     items,
			TotalSize: total,
		})
	}
	return groups, nil
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return formatInt(bytes) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	sizes := []string{"KB", "MB", "GB", "TB"}
	return formatFloat(float64(bytes)/float64(div)) + " " + sizes[exp]
}

func formatFloat(f float64) string {
	if f >= 100 {
		return formatInt(int64(f))
	}
	if f >= 10 {
		s := formatInt(int64(f * 10))
		return s[:len(s)-1] + "." + s[len(s)-1:]
	}
	s := formatInt(int64(f * 100))
	return s[:len(s)-2] + "." + s[len(s)-2:]
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

var electronCacheDirs = []struct {
	name string
	desc string
	risk RiskLevel
}{
	{"Cache", "Browser cache", RiskRegular},
	{"Code Cache", "JavaScript code cache", RiskRegular},
	{"GPUCache", "GPU rendering cache", RiskRegular},
	{"DawnGraphiteCache", "WebGPU cache", RiskRegular},
	{"DawnWebGPUCache", "WebGPU cache", RiskRegular},
	{"CachedData", "Cached data", RiskRegular},
	{"CachedProfilesData", "Cached profile data", RiskRegular},
	{"CachedConfigurations", "Cached configurations", RiskRegular},
	{"CachedExtensionVSIXs", "Cached extension packages", RiskRegular},
	{"Service Worker", "Service worker cache", RiskRegular},
	{"blob_storage", "Blob storage cache", RiskRegular},
	{"SharedClientCache", "Shared client cache", RiskRegular},
	{"Crashpad", "Crash reports", RiskRegular},
	{"CrashReport", "Crash reports", RiskRegular},
	{"logs", "Application logs", RiskRegular},
	{"Local Storage", "Local storage (login state)", RiskDeep},
	{"Session Storage", "Session storage", RiskDeep},
	{"Cookies", "Login cookies", RiskDeep},
	{"IndexedDB", "Local databases", RiskDeep},
	{"Network", "Network cache", RiskRegular},
}

func scanElectronDesktop(basePath string) []CacheItem {
	var items []CacheItem
	appSupport := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", basePath)
	cacheDir := filepath.Join(os.Getenv("HOME"), "Library", "Caches", basePath)
	for _, d := range electronCacheDirs {
		fullPath := filepath.Join(appSupport, d.name)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			size, _ := dirSize(fullPath)
			if size > 0 {
				items = append(items, CacheItem{
					Path:        fullPath,
					Size:        size,
					Description: d.desc,
					ModTime:     info.ModTime(),
					Risk:        d.risk,
				})
			}
		}
	}
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		size, _ := dirSize(cacheDir)
		if size > 0 {
			items = append(items, CacheItem{
				Path:        cacheDir,
				Size:        size,
				Description: "System cache directory",
				ModTime:     info.ModTime(),
			})
		}
	}
	return items
}
