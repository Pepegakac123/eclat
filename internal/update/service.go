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

// ReleaseInfo represents the information about a GitHub release.
type ReleaseInfo struct {
	TagName           string `json:"tagName"`
	Body              string `json:"body"`
	DownloadUrl       string `json:"downloadUrl"`
	IsUpdateAvailable bool   `json:"isUpdateAvailable"`
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

// CheckForUpdates checks GitHub for a newer version of the application.
func (s *UpdateService) CheckForUpdates() (ReleaseInfo, error) {
	s.logger.Info("Checking for updates...")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/pepegakac123/eclat/releases/latest")
	if err != nil {
		s.logger.Error("Failed to fetch latest release from GitHub", "error", err)
		return ReleaseInfo{}, fmt.Errorf("failed to fetch updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("GitHub API returned non-OK status", "status", resp.Status)
		return ReleaseInfo{}, fmt.Errorf("github api error: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		s.logger.Error("Failed to decode GitHub release info", "error", err)
		return ReleaseInfo{}, fmt.Errorf("failed to decode release info: %w", err)
	}

	remoteVer := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(release.TagName)), "v")
	currentVer := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(version.Version)), "v")

	isAvailable := remoteVer != "" && remoteVer != currentVer
	
	downloadUrl := ""
	// Find the appropriate asset based on the OS
	for _, asset := range release.Assets {
		if runtime.GOOS == "windows" && (filepath.Ext(asset.Name) == ".exe" || filepath.Ext(asset.Name) == ".msi") {
			downloadUrl = asset.BrowserDownloadUrl
			break
		}
		// For Mac/Linux we might just link to the release page or a generic package if we don't have a better strategy
	}

	// Fallback to browser download if no specific asset found or on non-windows
	if downloadUrl == "" {
		downloadUrl = "https://github.com/pepegakac123/eclat/releases/latest"
	}

	info := ReleaseInfo{
		TagName:           release.TagName,
		Body:              release.Body,
		DownloadUrl:       downloadUrl,
		IsUpdateAvailable: isAvailable,
	}

	s.logger.Info("Update check completed", "available", isAvailable, "version", release.TagName)
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

	// 2. Download
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

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

	// 3. Execute
	cmd := exec.Command(filePath)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start installer: %w", err)
	}

	s.logger.Info("Installer started, quitting application")
	
	// 4. Quit app to allow installer to overwrite
	go func() {
		time.Sleep(1 * time.Second)
		wailsRuntime.Quit(s.ctx)
	}()

	return "Installing... Application will close.", nil
}
