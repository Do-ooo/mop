package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	GitHubRepo    = "Do-ooo/mop"
	CheckInterval = 15 * 24 * time.Hour
)

var Version = "dev"

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type UpdateInfo struct {
	Available      bool
	LatestVersion  string
	DownloadURL    string
	CurrentVersion string
}

func CheckForUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "mop-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	info := &UpdateInfo{
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		Available:      latestVersion != currentVersion,
	}

	if info.Available {
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, runtime.GOOS) && strings.Contains(asset.Name, runtime.GOARCH) {
				info.DownloadURL = asset.BrowserDownloadURL
				break
			}
		}
		if info.DownloadURL == "" {
			info.Available = false
		}
	}

	return info, nil
}

func ShouldCheck() bool {
	configPath, err := updateCheckFile()
	if err != nil {
		return true
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return true
	}

	var lastCheck time.Time
	if err := lastCheck.UnmarshalText(data); err != nil {
		return true
	}

	return time.Since(lastCheck) >= CheckInterval
}

func RecordCheck() {
	configPath, err := updateCheckFile()
	if err != nil {
		return
	}

	data, _ := time.Now().MarshalText()
	os.WriteFile(configPath, data, 0644)
}

func updateCheckFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "mop")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "last_update_check"), nil
}

func DoUpdate() error {
	// mop self-update is macOS-only: releases only ship darwin binaries.
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("self-update is only supported on macOS")
	}

	elevated := false
	for _, a := range os.Args {
		if a == "--elevated" {
			elevated = true
		}
	}

	info, err := CheckForUpdate()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.Available {
		fmt.Println("Already up to date!")
		return nil
	}

	fmt.Printf("Updating from v%s to v%s...\n", info.CurrentVersion, info.LatestVersion)

	tmpPath, err := downloadBinary(info.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tmpPath)

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	if err := replaceBinary(currentExe, tmpPath); err != nil {
		// If the binary lives in a system directory we can't write to, re-exec
		// under sudo so the user gets a password prompt instead of a bare error.
		if isPermissionError(err) && !elevated {
			fmt.Println("Elevated permissions required to replace the binary. Re-running with sudo...")
			if reexecErr := reexecSudo(currentExe); reexecErr != nil {
				return fmt.Errorf("failed to replace binary: %w (try 'sudo mop update')", err)
			}
			return nil
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Updated to v%s!\n", info.LatestVersion)
	return nil
}

func isPermissionError(err error) bool {
	return errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "permission denied")
}

func reexecSudo(exePath string) error {
	cmd := exec.Command("sudo", exePath, "update", "--elevated")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func downloadBinary(url string) (string, error) {
	client := &http.Client{Timeout: 120 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "mop-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func replaceBinary(oldPath, newPath string) error {
	backupPath := oldPath + ".bak"

	if err := os.Rename(oldPath, backupPath); err != nil {
		return err
	}

	if err := os.Rename(newPath, oldPath); err != nil {
		os.Rename(backupPath, oldPath)
		return err
	}

	os.Remove(backupPath)
	return nil
}

var (
	updateMu         sync.RWMutex
	cachedUpdateInfo *UpdateInfo
	updateChecked    bool
)

func GetCachedUpdate() *UpdateInfo {
	updateMu.RLock()
	defer updateMu.RUnlock()
	if !updateChecked {
		return nil
	}
	return cachedUpdateInfo
}

func BackgroundCheck() {
	if !ShouldCheck() {
		updateMu.Lock()
		updateChecked = true
		updateMu.Unlock()
		return
	}

	go func() {
		info, err := CheckForUpdate()
		updateMu.Lock()
		if err == nil {
			cachedUpdateInfo = info
		}
		updateChecked = true
		updateMu.Unlock()
		if err == nil {
			RecordCheck()
		}
	}()
}
