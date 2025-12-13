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
var predefinedPalette = []struct {
	Name string
	Hex  string
}{
	// --- Szarości i Podstawowe ---
	{"Black", "#000000"},
	{"White", "#FFFFFF"},
	{"Dark Gray", "#404040"},
	{"Gray", "#808080"},
	{"Light Gray", "#C0C0C0"},

	// --- Czerwienie i Róże ---
	{"Dark Red", "#8B0000"},
	{"Red", "#FF0000"},
	{"Crimson", "#DC143C"}, // Karmazynowy
	{"Pink", "#FFC0CB"},
	{"Hot Pink", "#FF69B4"},
	{"Coral", "#FF7F50"},

	// --- Pomarańcze i Żółcie ---
	{"Brown", "#A52A2A"},
	{"Saddle Brown", "#8B4513"}, // Ciemny brąz (drewno)
	{"Orange", "#FFA500"},
	{"Gold", "#FFD700"},
	{"Yellow", "#FFFF00"},
	{"Beige", "#F5F5DC"}, // Beż (skóra/papier)

	// --- Zielenie ---
	{"Olive", "#808000"},
	{"Dark Green", "#006400"},
	{"Green", "#008000"},
	{"Lime", "#00FF00"},
	{"Teal", "#008080"}, // Morski ciemny

	// --- Niebieskie i Cyjany ---
	{"Cyan", "#00FFFF"},
	{"Sky Blue", "#87CEEB"},
	{"Blue", "#0000FF"},
	{"Navy", "#000080"}, // Granatowy
	{"Turquoise", "#40E0D0"},

	// --- Fiolety ---
	{"Indigo", "#4B0082"},
	{"Purple", "#800080"},
	{"Violet", "#EE82EE"},
	{"Lavender", "#E6E6FA"},
	{"Magenta", "#FF00FF"},
}

// ScannerConfig holds the configuration settings for the file scanner.
// It determines which files are processed based on their extensions.
type ScannerConfig struct {
	AllowedExtensions map[string]bool `json:"allowedExtensions"`
	PredefinedPalette []struct {
		Name string
		Hex  string
	} `json:"predefinedPalette"`
	MaxAllowHashFileSize int64 `json:"maxAllowHashFileSize"`
}

// NewScannerConfig initializes a new configuration with default allowed extensions.
// It performs a deep copy of the default map to prevent global state mutation.
func NewScannerConfig() *ScannerConfig {
	defaultAllowedExtensionsCopy := make(map[string]bool)
	maps.Copy(defaultAllowedExtensionsCopy, defaultAllowedExtensions)
	return &ScannerConfig{
		AllowedExtensions:    defaultAllowedExtensionsCopy,
		PredefinedPalette:    predefinedPalette,
		MaxAllowHashFileSize: 1024 * 1024 * 10, // 10MB
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
		PredefinedPalette: predefinedPalette,
	}
}

// GetColorPalette returns the hex values of the default color palette.
func (s *Scanner) GetColorsPaletteHex() []string {
	palette := []string{}
	for _, color := range s.config.PredefinedPalette {
		if color.Hex != "" {
			palette = append(palette, color.Hex)
		}
	}
	return palette
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
