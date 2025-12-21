package config

import (
	"path/filepath"
	"slices"
	"strings"
)

type PaletteColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// Exported variables (Capitalized)
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
	AllowedExtensions    []string `json:"allowedExtensions"`
	MaxAllowHashFileSize int64    `json:"maxAllowHashFileSize"`
}

func NewScannerConfig() *ScannerConfig {
	exts := make([]string, len(DefaultAllowedExtensions))
	copy(exts, DefaultAllowedExtensions)

	return &ScannerConfig{
		AllowedExtensions:    exts,
		MaxAllowHashFileSize: 1024 * 1024 * 256,
	}
}

// IsExtensionValid checks if extension is safe. Moved here as a helper.
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
