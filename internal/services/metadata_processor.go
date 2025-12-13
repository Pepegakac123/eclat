package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/tiff"

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

type ImageMetadata struct {
	Width           int
	Height          int
	HasAlphaChannel bool
	BitDepth        int
}

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

func (s *Scanner) extractImageMetadata(filepath string) (ImageMetadata, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return ImageMetadata{}, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return ImageMetadata{}, err
	}

	meta := ImageMetadata{
		Width:  cfg.Width,
		Height: cfg.Height,
	}

	model := cfg.ColorModel
	// Sprawdzenie kanału Alpha
	// JPEG nie ma alfy, PNG/GIF/WebP mogą mieć.
	// Sprawdzamy, czy model to jeden z typów wspierających przezroczystość.
	switch model {
	case color.RGBAModel, color.NRGBAModel, color.AlphaModel, color.Alpha16Model, color.NYCbCrAModel:
		meta.HasAlphaChannel = true
	default:
		// Bardziej zaawansowane sprawdzenie dla palet (np. GIF/PNG8)
		if _, ok := model.(color.Palette); ok {
			// Palety mogą mieć przezroczystość, ale decodeConfig tego łatwo nie powie bez analizy palety.
			// Dla uproszczenia w MVP: GIFy często mają alpha.
			meta.HasAlphaChannel = true
		} else {
			meta.HasAlphaChannel = false
		}
	}

	// Sprawdzenie głębi bitowej (Bit Depth)
	switch model {
	case color.RGBA64Model, color.NRGBA64Model, color.Alpha16Model, color.Gray16Model:
		meta.BitDepth = 16
	default:
		meta.BitDepth = 8
	}

	return meta, nil
}
func (s *Scanner) calculateFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fileInfo.Size() > s.config.MaxAllowHashFileSize {
		return "", errors.New("file size exceeds maximum allowed for it to be hashed")
	}

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
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
