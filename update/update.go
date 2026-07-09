package update

import (
	"bufio"
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
	GitHubRepo = "Do-ooo/mop"
	// RawVersionURL points at the plain-text VERSION file on the main branch.
	// It is served by GitHub's raw CDN, which is NOT rate-limited like the API,
	// so update checks never hit the 60 req/hour/IP cap that caused 403s.
	RawVersionURL = "https://raw.githubusercontent.com/Do-ooo/mop/main/VERSION"
	CheckInterval = 15 * 24 * time.Hour
)

var Version = "dev"

type UpdateInfo struct {
	Available      bool
	LatestVersion  string
	DownloadURL    string
	CurrentVersion string
}

func CheckForUpdate() (*UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", RawVersionURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mop-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("version check returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	latestVersion := strings.TrimPrefix(strings.TrimSpace(string(body)), "v")
	if latestVersion == "" {
		return nil, fmt.Errorf("empty version file")
	}
	currentVersion := strings.TrimPrefix(Version, "v")

	info := &UpdateInfo{
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		Available:      latestVersion != currentVersion,
	}

	// Release asset names follow a fixed scheme (mop-<os>-<arch>), so the
	// download URL can be derived from the version alone — no API call needed.
	if info.Available {
		info.DownloadURL = fmt.Sprintf(
			"https://github.com/%s/releases/download/v%s/mop-%s-%s",
			GitHubRepo, latestVersion, runtime.GOOS, runtime.GOARCH,
		)
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

// StartupCheck runs before the TUI launches. When a periodic check is due it
// queries the latest version synchronously — cheap now that it only fetches a
// few-byte VERSION file from the raw CDN — and, if a newer version exists,
// prompts the user in the bare terminal (oh-my-zsh style) before entering the
// full-screen UI. Returns true if an update was performed, in which case the
// caller should exit instead of launching the TUI.
//
// A failed check degrades silently: startup is never blocked by network issues.
func StartupCheck() bool {
	if !ShouldCheck() {
		return false
	}

	info, err := CheckForUpdate()
	if err != nil {
		return false
	}
	RecordCheck()

	// Cache the result so the in-TUI banner can remind the user if they decline.
	updateMu.Lock()
	cachedUpdateInfo = info
	updateChecked = true
	updateMu.Unlock()

	if !info.Available {
		return false
	}

	fmt.Printf("New version available: v%s (current v%s)\n", info.LatestVersion, info.CurrentVersion)
	fmt.Print("Update now? [Y/n] ")

	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "n", "no":
		return false
	}

	// Default (empty / y / yes) proceeds with the update.
	if err := DoUpdate(); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return false
	}
	return true
}
