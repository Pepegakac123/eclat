package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
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

type ImageMetadata struct {
	Width           int
	Height          int
	HasAlphaChannel bool
	BitDepth        int
	DominantColor   string
}

func DetermineFileType(extension string) string {
	ext := strings.ToLower(extension)

	switch ext {
	// Images
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return string(FileTypeImage)

	// 3D Models
	case ".blend", ".blend1", ".fbx", ".obj", ".max", ".ma", ".mb", ".glb", ".gltf":
		return string(FileTypeModel)

	// Sculpting
	case ".ztl", ".zpr", ".zbr":
		return string(FileTypeModel)

	// Procedural
	case ".hip", ".hipnc", ".hiplc":
		return string(FileTypeModel)

	// Game Engine
	case ".uasset", ".umap", ".unity", ".prefab", ".mat", ".asset":
		return string(FileTypeModel)

	// Textures
	case ".psd", ".psb", ".ai", ".eps", ".exr", ".hdr", ".tif", ".tiff", ".tga":
		return string(FileTypeTexture)

	// Substance
	case ".spp", ".sbs", ".sbsar":
		return string(FileTypeTexture)

	default:
		return string(FileTypeOther)
	}
}

func CalculateFileHash(filePath string, maxSizeBytes int64) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	if maxSizeBytes > 0 && fileInfo.Size() > maxSizeBytes {
		return "", errors.New("file size exceeds maximum allowed for hashing")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CalculateDominantColor(img image.Image) (string, error) {
	if img == nil {
		return "", errors.New("image is nil")
	}

	resizeSize := uint(prominentcolor.DefaultSize)
	bgmasks := prominentcolor.GetDefaultMasks()
	k := 1 // Chcemy jeden główny kolor
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

func FindClosestPaletteColor(hexInput string, palette []struct {
	Name string
	Hex  string
}) (string, error) {
	inputColor, err := colorful.Hex(hexInput)
	if err != nil {
		return "", fmt.Errorf("invalid hex color: %w", err)
	}

	var closestHex string
	minDistance := math.MaxFloat64

	for _, paletteItem := range palette {
		pColor, err := colorful.Hex(paletteItem.Hex)
		if err != nil {
			continue
		}

		// DistanceLab (CIELAB)
		dist := inputColor.DistanceLab(pColor)

		if dist < minDistance {
			minDistance = dist
			closestHex = paletteItem.Hex
		}
	}

	return closestHex, nil
}

// GetBitDepth analizuje model kolorów obrazu i zwraca głębię bitową na kanał.
// Zwraca 8 (standard), 16 (High Quality) lub inne wartości.
func GetBitDepth(img image.Image) int {
	if img == nil {
		return 0
	}

	switch img.ColorModel() {
	// Modele 16-bitowe (High Dynamic Range / RAW exports)
	case color.RGBA64Model, color.NRGBA64Model, color.Alpha16Model, color.Gray16Model:
		return 16

	// Modele CMYK (rzadkie w web, ale możliwe w druku) - zazwyczaj 8 bit na kanał
	case color.CMYKModel:
		return 8

	// Standardowe 8-bitowe
	case color.RGBAModel, color.NRGBAModel, color.AlphaModel, color.GrayModel:
		return 8

	// Palety (GIF, PNG-8) - technicznie to indeksy (często < 8 bit),
	default:
		// Sprawdźmy czy to paleta
		if _, ok := img.ColorModel().(color.Palette); ok {
			return 8
		}
		// Fallback dla nietypowych formatów
		return 8
	}
}
