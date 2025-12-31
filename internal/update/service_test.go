package update

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeVersion(t *testing.T) {
	assert.Equal(t, "1.2.3", normalizeVersion("v1.2.3"))
	assert.Equal(t, "1.2.3", normalizeVersion(" 1.2.3 "))
	assert.Equal(t, "1.2.3", normalizeVersion("V1.2.3"))
}

func TestAssetSelection(t *testing.T) {
	// Mock releases
	releases := []GitHubRelease{
		{
			TagName: "v1.1.0",
			Assets: []struct {
				Name               string `json:"name"`
				BrowserDownloadUrl string `json:"browser_download_url"`
			}{
				{"eclat-amd64-installer.exe", "url-amd64-installer"},
				{"eclat-amd64.exe", "url-amd64"},
				{"eclat-linux-amd64", "url-linux"},
				{"eclat-windows-installer.exe", "url-windows-installer"},
			},
		},
	}

	// This is a simplified version of the logic in CheckForUpdates
	findBestAsset := func(assets []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}, goos, goarch string) string {
		var bestAsset string
		maxPriority := -1

		for _, asset := range assets {
			name := strings.ToLower(asset.Name)
			ext := ""
			if len(name) > 4 {
				ext = name[len(name)-4:]
			}

			if goos == "windows" && (ext == ".exe" || ext == ".msi") {
				priority := 0
				if strings.Contains(name, "installer") || strings.Contains(name, "setup") {
					priority += 10
				}
				if strings.Contains(name, "windows") {
					priority += 5
				}
				if goarch == "amd64" && (strings.Contains(name, "amd64") || strings.Contains(name, "x64")) {
					priority += 3
				}

				if priority > maxPriority {
					maxPriority = priority
					bestAsset = asset.BrowserDownloadUrl
				}
			}
		}
		return bestAsset
	}

	url := findBestAsset(releases[0].Assets, "windows", "amd64")
	// Should prefer "eclat-windows-installer.exe" or "eclat-amd64-installer.exe"
	// eclat-windows-installer.exe -> installer(10) + windows(5) = 15
	// eclat-amd64-installer.exe -> installer(10) + amd64(3) = 13
	assert.Equal(t, "url-windows-installer", url)
}
