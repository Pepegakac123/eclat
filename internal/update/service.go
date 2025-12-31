package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"eclat/internal/version"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ReleaseNote represents a single version's changelog.
type ReleaseNote struct {
	TagName string `json:"tagName"`
	Body    string `json:"body"`
}

// ReleaseInfo represents the information about a GitHub release.
type ReleaseInfo struct {
	TagName           string        `json:"tagName"`
	Body              string        `json:"body"`
	DownloadUrl       string        `json:"downloadUrl"`
	IsUpdateAvailable bool          `json:"isUpdateAvailable"`
	History           []ReleaseNote `json:"history"`
}

// GitHubRelease is the structure for parsing GitHub API response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpdateService handles checking and installing application updates.
type UpdateService struct {
	ctx    context.Context
	logger *slog.Logger
}

// NewUpdateService creates a new UpdateService instance.
func NewUpdateService(logger *slog.Logger) *UpdateService {
	return &UpdateService{
		logger: logger,
	}
}

// Startup is called at application startup.
func (s *UpdateService) Startup(ctx context.Context) {
	s.ctx = ctx
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(v)), "v")
}

// CheckForUpdates checks GitHub for newer versions of the application.
func (s *UpdateService) CheckForUpdates() (ReleaseInfo, error) {
	s.logger.Info("Checking for updates...")
	client := &http.Client{Timeout: 10 * time.Second}
	// Fetch all releases instead of just latest to build history
	resp, err := client.Get("https://api.github.com/repos/pepegakac123/eclat/releases")
	if err != nil {
		s.logger.Error("Failed to fetch releases from GitHub", "error", err)
		return ReleaseInfo{}, fmt.Errorf("failed to fetch updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("GitHub API returned non-OK status", "status", resp.Status)
		return ReleaseInfo{}, fmt.Errorf("github api error: %s", resp.Status)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		s.logger.Error("Failed to decode GitHub releases info", "error", err)
		return ReleaseInfo{}, fmt.Errorf("failed to decode releases info: %w", err)
	}

	if len(releases) == 0 {
		return ReleaseInfo{IsUpdateAvailable: false}, nil
	}

	currentVer := normalizeVersion(version.Version)
	latestRelease := releases[0]
	remoteVer := normalizeVersion(latestRelease.TagName)

	isAvailable := remoteVer != "" && remoteVer != currentVer

	var history []ReleaseNote
	downloadUrl := ""

	// Build history of releases newer than current version
	for _, rel := range releases {
		relVer := normalizeVersion(rel.TagName)
		if relVer == currentVer {
			break // Reached current version, stop history
		}
		
		history = append(history, ReleaseNote{
			TagName: rel.TagName,
			Body:    rel.Body,
		})

		// Use download URL from the very first (latest) release
		if downloadUrl == "" {
			var bestAsset string
			maxPriority := -1

			for _, asset := range rel.Assets {
				name := strings.ToLower(asset.Name)
				ext := filepath.Ext(name)

				if runtime.GOOS == "windows" && (ext == ".exe" || ext == ".msi") {
					priority := 0
					// Prefer installers
					if strings.Contains(name, "installer") || strings.Contains(name, "setup") {
						priority += 10
					}
					// Prefer windows-specific names if multiple exe exist
					if strings.Contains(name, "windows") {
						priority += 5
					}
					// Match architecture
					arch := runtime.GOARCH
					if arch == "amd64" && (strings.Contains(name, "amd64") || strings.Contains(name, "x64")) {
						priority += 3
					} else if arch == "arm64" && strings.Contains(name, "arm64") {
						priority += 3
					}

					if priority > maxPriority {
						maxPriority = priority
						bestAsset = asset.BrowserDownloadUrl
					}
				}
			}
			downloadUrl = bestAsset
		}
	}

	// Fallback to browser download if no specific asset found or on non-windows
	if downloadUrl == "" {
		downloadUrl = "https://github.com/pepegakac123/eclat/releases/latest"
	}

	info := ReleaseInfo{
		TagName:           latestRelease.TagName,
		Body:              latestRelease.Body,
		DownloadUrl:       downloadUrl,
		IsUpdateAvailable: isAvailable,
		History:           history,
	}

	s.logger.Info("Update check completed", "available", isAvailable, "version", latestRelease.TagName, "history_count", len(history))
	return info, nil
}

// DownloadAndInstall handles the download and execution of the update.
func (s *UpdateService) DownloadAndInstall(url string) (string, error) {
	if runtime.GOOS != "windows" {
		s.logger.Info("Non-windows update requested, opening browser", "url", url)
		wailsRuntime.BrowserOpenURL(s.ctx, url)
		return "Opened download page in browser.", nil
	}

	s.logger.Info("Starting Windows update download", "url", url)
	
	// 1. Create temp file
	tempDir := os.TempDir()
	fileName := fmt.Sprintf("eclat-update-%d.exe", time.Now().Unix())
	filePath := filepath.Join(tempDir, fileName)

	// 2. Download with User-Agent
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "eclat-updater")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned non-OK status: %s", resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save update file: %w", err)
	}

	s.logger.Info("Update downloaded successfully", "path", filePath)

	// 3. Execute - Use cmd /c start to detach and handle elevation properly
	cmd := exec.Command("cmd", "/c", "start", "", filePath)
	if err := cmd.Run(); err != nil {
		s.logger.Error("Failed to start installer via cmd", "error", err)
		// Fallback to direct execution
		cmdFallback := exec.Command(filePath)
		if err := cmdFallback.Start(); err != nil {
			return "", fmt.Errorf("failed to start installer: %w", err)
		}
	}

	s.logger.Info("Installer started, quitting application")
	
	// 4. Quit app to allow installer to overwrite
	go func() {
		// Wait a bit longer to ensure installer UI is visible
		time.Sleep(2 * time.Second)
		wailsRuntime.Quit(s.ctx)
	}()

	return "Installing... Application will close shortly.", nil
}
