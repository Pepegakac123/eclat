package services

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

var defaultAllowedExtensions = map[string]bool{
	".jpg":    true,
	".jpeg":   true,
	".gif":    true,
	".png":    true,
	".webp":   true,
	".blend":  true,
	".fbx":    true,
	".obj":    true,
	".ztl":    true,
	".zpr":    true,
	".exr":    true,
	".hdr":    true,
	".tif":    true,
	".tiff":   true,
	".max":    true,
	".ma":     true,
	".mb":     true,
	".zbr":    true,
	".spp":    true,
	".sbs":    true,
	".sbsar":  true,
	".hip":    true,
	".hipnc":  true,
	".hiplc":  true,
	".psd":    true,
	".psb":    true,
	".ai":     true,
	".eps":    true,
	".uasset": true,
	".umap":   true,
	".unity":  true,
	".prefab": true,
	".mat":    true,
	".asset":  true,
}

var dangerousExtensions = []string{".exe", ".dll", ".bat", ".cmd", ".sh", ".vbs", ".msi", ".com", ".scr", ".js", ".ps1", ".bin"}

// ScannerConfig holds the configuration settings for the file scanner.
// It determines which files are processed based on their extensions.
type ScannerConfig struct {
	AllowedExtensions map[string]bool `json:"allowedExtensions"`
}

// NewScannerConfig initializes a new configuration with default allowed extensions.
// It performs a deep copy of the default map to prevent global state mutation.
func NewScannerConfig() *ScannerConfig {
	defaultAllowedExtensionsCopy := make(map[string]bool)
	maps.Copy(defaultAllowedExtensionsCopy, defaultAllowedExtensions)
	return &ScannerConfig{
		AllowedExtensions: defaultAllowedExtensionsCopy,
	}
}

// GetConfig returns a thread-safe DEEP COPY of the current configuration.
// Returning a copy ensures that the caller cannot modify the internal state of the Scanner
// and avoids race conditions when reading the map outside the mutex lock.
func (s *Scanner) GetConfig() ScannerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	safeMap := make(map[string]bool, len(s.config.AllowedExtensions))
	maps.Copy(safeMap, s.config.AllowedExtensions)

	return ScannerConfig{
		AllowedExtensions: safeMap,
	}
}

// isExtensionValid is a helper method that checks if the provided extension is valid and not dangerous
func isExtensionValid(ext string) bool {
	f := strings.ToLower(filepath.Ext(ext))
	if !strings.HasPrefix(f, ".") {
		f = "." + f
	}
	if len(f) <= 1 {
		return false
	}
	return !slices.Contains(dangerousExtensions, f)
}

// IsExtensionAllowed checks if the extension is in the allowed extensions to scan map
func (s *Scanner) IsExtensionAllowed(ext string) bool {
	normalized := strings.ToLower(ext)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.AllowedExtensions[normalized]
}

// AddExtensions adds new allowed extensions to the configuration.
// It returns an aggregate error if any of the provided extensions are invalid.
// Valid extensions are added even if some others fail.
func (s *Scanner) AddExtensions(exts []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var invalidExts = []string{}
	if len(exts) <= 0 {
		return nil
	}
	for _, ext := range exts {
		normalized := strings.ToLower(ext)
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}

		if !isExtensionValid(normalized) {
			invalidExts = append(invalidExts, ext)
			continue
		}
		if !s.config.AllowedExtensions[normalized] {
			s.config.AllowedExtensions[normalized] = true
		}
	}
	if len(invalidExts) > 0 {
		return fmt.Errorf("invalid or dangerous extensions detected: %s", strings.Join(invalidExts, ", "))
	}
	return nil
}

// RemoveExtension removes a specific extension from the allowed list.
func (s *Scanner) RemoveExtension(ext string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalized := strings.ToLower(ext)
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}
	delete(s.config.AllowedExtensions, normalized)
}
