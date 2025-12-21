package services

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type PaletteColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// Domyślne rozszerzenia jako slice
var defaultAllowedExtensions = []string{
	".jpg", ".jpeg", ".gif", ".png", ".webp", ".blend", ".fbx", ".obj",
	".ztl", ".zpr", ".exr", ".hdr", ".tif", ".tiff", ".max", ".ma", ".mb",
	".zbr", ".spp", ".sbs", ".sbsar", ".hip", ".hipnc", ".hiplc", ".psd",
	".psb", ".ai", ".eps", ".uasset", ".umap", ".unity", ".prefab", ".mat", ".asset",
}

var dangerousExtensions = []string{".exe", ".dll", ".bat", ".cmd", ".sh", ".vbs", ".msi", ".com", ".scr", ".js", ".ps1", ".bin"}

var predefinedPalette = []PaletteColor{
	{"Black", "#000000"}, {"White", "#FFFFFF"}, {"Dark Gray", "#404040"}, {"Gray", "#808080"}, {"Light Gray", "#C0C0C0"},
	{"Dark Red", "#8B0000"}, {"Red", "#FF0000"}, {"Crimson", "#DC143C"}, {"Pink", "#FFC0CB"}, {"Hot Pink", "#FF69B4"}, {"Coral", "#FF7F50"},
	{"Brown", "#A52A2A"}, {"Saddle Brown", "#8B4513"}, {"Orange", "#FFA500"}, {"Gold", "#FFD700"}, {"Yellow", "#FFFF00"}, {"Beige", "#F5F5DC"},
	{"Olive", "#808000"}, {"Dark Green", "#006400"}, {"Green", "#008000"}, {"Lime", "#00FF00"}, {"Teal", "#008080"},
	{"Cyan", "#00FFFF"}, {"Sky Blue", "#87CEEB"}, {"Blue", "#0000FF"}, {"Navy", "#000080"}, {"Turquoise", "#40E0D0"},
	{"Indigo", "#4B0082"}, {"Purple", "#800080"}, {"Violet", "#EE82EE"}, {"Lavender", "#E6E6FA"}, {"Magenta", "#FF00FF"},
}

// ScannerConfig - DTO dla Frontendu.
// WAŻNE: Używamy []string zamiast mapy, aby uniknąć błędów Signal 11 w Wails/CGO.
type ScannerConfig struct {
	AllowedExtensions    []string `json:"allowedExtensions"`
	MaxAllowHashFileSize int64    `json:"maxAllowHashFileSize"`
}

// NewScannerConfig tworzy domyślną konfigurację.
func NewScannerConfig() *ScannerConfig {
	// Kopiujemy domyślne rozszerzenia do nowej listy
	exts := make([]string, len(defaultAllowedExtensions))
	copy(exts, defaultAllowedExtensions)

	return &ScannerConfig{
		AllowedExtensions:    exts,
		MaxAllowHashFileSize: 1024 * 1024 * 256, //
	}
}

// GetConfig zwraca kopię konfiguracji bezpieczną dla wątków.
func (s *Scanner) GetConfig() ScannerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Kopiujemy slice, aby frontend nie mógł modyfikować oryginału
	safeExts := make([]string, len(s.config.AllowedExtensions))
	copy(safeExts, s.config.AllowedExtensions)

	return ScannerConfig{
		AllowedExtensions:    safeExts,
		MaxAllowHashFileSize: s.config.MaxAllowHashFileSize,
	}
}

// GetPredefinedPalette zwraca paletę kolorów.
func (s *Scanner) GetPredefinedPalette() []PaletteColor {
	return predefinedPalette
}

// IsExtensionAllowed sprawdza, czy plik może być skanowany.
func (s *Scanner) IsExtensionAllowed(ext string) bool {
	normalized := strings.ToLower(ext)
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sprawdzamy w liście (slice). Dla <100 elementów jest to błyskawiczne.
	return slices.Contains(s.config.AllowedExtensions, normalized)
}

// AddExtensions dodaje nowe rozszerzenia.
func (s *Scanner) AddExtensions(exts []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var invalidExts []string
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

		// Dodajemy tylko jeśli nie istnieje
		if !slices.Contains(s.config.AllowedExtensions, normalized) {
			s.config.AllowedExtensions = append(s.config.AllowedExtensions, normalized)
		}
	}

	if len(invalidExts) > 0 {
		return fmt.Errorf("invalid or dangerous extensions: %s", strings.Join(invalidExts, ", "))
	}
	return nil
}

// RemoveExtension usuwa rozszerzenie z listy.
func (s *Scanner) RemoveExtension(ext string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(ext)
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}

	// Usuwamy element ze slice'a
	s.config.AllowedExtensions = slices.DeleteFunc(s.config.AllowedExtensions, func(e string) bool {
		return e == normalized
	})
}

// Helpery
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
