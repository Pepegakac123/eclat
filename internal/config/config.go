package config

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type PaletteColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

var DefaultAllowedExtensions = []string{
	".jpg", ".jpeg", ".gif", ".png", ".webp", ".blend", ".fbx", ".obj",
	".ztl", ".zpr", ".exr", ".hdr", ".tif", ".tiff", ".max", ".ma", ".mb",
	".zbr", ".spp", ".sbs", ".sbsar", ".hip", ".hipnc", ".hiplc", ".psd",
	".psb", ".ai", ".eps", ".uasset", ".umap", ".unity", ".prefab", ".mat", ".asset",
}

var DangerousExtensions = []string{".exe", ".dll", ".bat", ".cmd", ".sh", ".vbs", ".msi", ".com", ".scr", ".js", ".ps1", ".bin"}

var PredefinedPalette = []PaletteColor{
	{"Black", "#000000"}, {"White", "#FFFFFF"}, {"Dark Gray", "#404040"}, {"Gray", "#808080"}, {"Light Gray", "#C0C0C0"},
	{"Dark Red", "#8B0000"}, {"Red", "#FF0000"}, {"Crimson", "#DC143C"}, {"Pink", "#FFC0CB"}, {"Hot Pink", "#FF69B4"}, {"Coral", "#FF7F50"},
	{"Brown", "#A52A2A"}, {"Saddle Brown", "#8B4513"}, {"Orange", "#FFA500"}, {"Gold", "#FFD700"}, {"Yellow", "#FFFF00"}, {"Beige", "#F5F5DC"},
	{"Olive", "#808000"}, {"Dark Green", "#006400"}, {"Green", "#008000"}, {"Lime", "#00FF00"}, {"Teal", "#008080"},
	{"Cyan", "#00FFFF"}, {"Sky Blue", "#87CEEB"}, {"Blue", "#0000FF"}, {"Navy", "#000080"}, {"Turquoise", "#40E0D0"},
	{"Indigo", "#4B0082"}, {"Purple", "#800080"}, {"Violet", "#EE82EE"}, {"Lavender", "#E6E6FA"}, {"Magenta", "#FF00FF"},
}

type ScannerConfig struct {
	allowedExtensions    []string
	maxAllowHashFileSize int64
	mu                   sync.RWMutex
}

func NewScannerConfig() *ScannerConfig {
	exts := make([]string, len(DefaultAllowedExtensions))
	copy(exts, DefaultAllowedExtensions)

	return &ScannerConfig{
		allowedExtensions:    exts,
		maxAllowHashFileSize: 1024 * 1024 * 256,
	}
}

func (c *ScannerConfig) GetAllowedExtensions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.allowedExtensions))
	copy(result, c.allowedExtensions)
	return result
}

// SetAllowedExtensions bezpiecznie podmienia listę
func (c *ScannerConfig) SetAllowedExtensions(exts []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	newExts := make([]string, len(exts))
	copy(newExts, exts)
	c.allowedExtensions = newExts
}

func (c *ScannerConfig) GetMaxHashFileSize() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxAllowHashFileSize
}

// IsExtensionAllowed sprawdza czy plik ma dozwolone rozszerzenie (sprawdza w konfigu instancji)
func (c *ScannerConfig) IsExtensionAllowed(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(c.allowedExtensions, ext)
}

func IsExtensionValid(ext string) bool {
	f := strings.ToLower(ext)
	if strings.Contains(f, ".") {
		f = filepath.Ext(f)
	}

	if !strings.HasPrefix(f, ".") {
		f = "." + f
	}

	if len(f) <= 1 {
		return false
	}
	return !slices.Contains(DangerousExtensions, f)
}
