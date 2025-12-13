package services

import (
	"database/sql"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/lucasb-eyer/go-colorful"
)

type FileType string

const (
	FileTypeImage   FileType = "image"
	FileTypeModel   FileType = "model"
	FileTypeTexture FileType = "texture"
	FileTypeOther   FileType = "other"
)

// determineFileType przypisuje kategorię na podstawie rozszerzenia pliku.
// Zakłada, że extension zawiera kropkę (np. ".jpg").
func (s *Scanner) determineFileType(extension string) string {
	ext := strings.ToLower(extension)

	switch ext {
	// Images
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return string(FileTypeImage)

	// 3D Models (Blender, FBX, OBJ, Maya, Max)
	case ".blend", ".blend1", ".fbx", ".obj", ".max", ".ma", ".mb":
		return string(FileTypeModel)

	// Sculpting (ZBrush)
	case ".ztl", ".zpr", ".zbr":
		return string(FileTypeModel)

	// Procedural/Houdini
	case ".hip", ".hipnc", ".hiplc":
		return string(FileTypeModel)

	// Game Engine Assets (Unreal, Unity)
	case ".uasset", ".umap", ".unity", ".prefab", ".mat", ".asset":
		return string(FileTypeModel)

	// Textures/Materials (Adobe, EXR, TIFF)
	case ".psd", ".psb", ".ai", ".eps", ".exr", ".hdr", ".tif", ".tiff":
		return string(FileTypeTexture)

	// Substance (Adobe Substance)
	case ".spp", ".sbs", ".sbsar":
		return string(FileTypeTexture)

	default:
		return string(FileTypeOther)
	}
}

// calculateDominantColor calculates the dominant color hex string using K-means clustering.
// It analyzes the specific file path.
func calculateDominantColor(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// TODO: Performance warning - to wczytuje pełną rozdzielczość do RAM.
	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	//  Konfiguracja K-means
	// K = 1 (chcemy tylko jeden dominujący kolor)
	// ArgumentNoCropping (analizuj całość, nie tylko środek)

	resizeSize := uint(prominentcolor.DefaultSize)
	bgmasks := prominentcolor.GetDefaultMasks()
	k := 1
	centroids, err := prominentcolor.KmeansWithAll(
		k,
		img,
		prominentcolor.ArgumentNoCropping,
		resizeSize,
		bgmasks,
	)

	if err != nil {
		return "", fmt.Errorf("kmeans processing failed: %w", err)
	}

	if len(centroids) == 0 {
		return "", fmt.Errorf("no dominant color found")
	}
	c := centroids[0]
	hexColor := fmt.Sprintf("#%02X%02X%02X", c.Color.R, c.Color.G, c.Color.B)

	return hexColor, nil
}

// findClosestPaletteColor finds the closest color from the predefined palette for a given hex string.
// It uses the CIE L*a*b* color space for the best perceptual precision.
func findClosestPaletteColor(hexInput string, pallete []struct {
	Name string
	Hex  string
}) (string, error) {
	inputColor, err := colorful.Hex(hexInput)
	if err != nil {
		return "", fmt.Errorf("invalid hex color: %w", err)
	}

	var closestHex string
	minDistance := math.MaxFloat64
	for _, paletteItem := range pallete {
		pColor, err := colorful.Hex(paletteItem.Hex)
		if err != nil {
			continue
		}
		dist := inputColor.DistanceLab(pColor)

		if dist < minDistance {
			minDistance = dist
			closestHex = paletteItem.Hex
		}
	}

	return closestHex, nil
}

func (s *Scanner) getDominantColor(path string) sql.NullString {
	ext := filepath.Ext(path)
	fileType := s.determineFileType(ext)
	if fileType != string(FileTypeImage) {
		return sql.NullString{}
	}
	dominantColor, err := calculateDominantColor(path)
	if err != nil {
		s.logger.Debug("Failed to calc dominant color", "file", path, "error", err)
		return sql.NullString{}
	}
	closestColor, err := findClosestPaletteColor(dominantColor, s.config.PredefinedPalette)
	if err != nil {
		s.logger.Debug("Failed to normalize color to palette", "file", path, "error", err)
		return sql.NullString{}
	}
	return sql.NullString{String: closestColor, Valid: true}

}
