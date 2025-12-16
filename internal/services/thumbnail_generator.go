package services

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"path"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type ThumbnailGenerator struct {
	cacheDir       string
	logger         *slog.Logger
	placeholderMap map[string]string
}
type ThumbnailResult struct {
	WebPath       string
	Metadata      ImageMetadata
	IsPlaceholder bool
}

func NewThumbnailGenerator(cacheDir string, logger *slog.Logger) *ThumbnailGenerator {
	return &ThumbnailGenerator{
		cacheDir: cacheDir,
		logger:   logger,
		placeholderMap: map[string]string{
			".blend":  "blend_placeholder.webp",
			".blend1": "blend_placeholder.webp",

			".max": "max_placeholder.webp",

			".ma": "ma_placeholder.webp",
			".mb": "ma_placeholder.webp",

			".ztl": "ztl_placeholder.webp",
			".zpr": "ztl_placeholder.webp",
			".zbr": "ztl_placeholder.webp",

			".spp":   "spp_placeholder.webp",
			".sbs":   "sbs_placeholder.webp",
			".sbsar": "sbs_placeholder.webp",

			".hip":   "hip_placeholder.webp",
			".hipnc": "hip_placeholder.webp",
			".hiplc": "hip_placeholder.webp",

			".psd": "psd_placeholder.webp",
			".psb": "psd_placeholder.webp",
			".ai":  "ai_placeholder.webp",
			".eps": "ai_placeholder.webp",

			".uasset": "uasset_placeholder.webp",
			".umap":   "uasset_placeholder.webp",
			".unity":  "unity_placeholder.webp",
			".prefab": "unity_placeholder.webp",
			".mat":    "unity_placeholder.webp",
			".asset":  "unity_placeholder.webp",
		},
	}
}

func isSupportedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tif", ".tiff":
		return true
	}
	return false
}

func (g *ThumbnailGenerator) Generate(ctx context.Context, srcPath string) (ThumbnailResult, error) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if isSupportedImageExt(ext) {
		return g.generateFromImage(srcPath)
	}
	return g.getPlaceholderResult(ext), nil
}
func (g *ThumbnailGenerator) generateFromImage(srcPath string) (ThumbnailResult, error) {
	img, err := imaging.Open(srcPath)
	if err != nil {
		g.logger.Warn("Failed to decode image, using placeholder", "path", srcPath, "error", err)
		return g.getPlaceholderResult(filepath.Ext(srcPath)), nil
	}
	metadata := g.extractMetadataFromImage(img)

	// 3. Generowanie miniatury
	id := uuid.New()
	filename := fmt.Sprintf("%s.webp", id.String())
	fullDestPath := filepath.Join(g.cacheDir, filename)

	// Resize (szybki Lanczos)
	thumb := imaging.Resize(img, 400, 0, imaging.Lanczos)

	// Zapis
	if err := imaging.Save(thumb, fullDestPath); err != nil {
		return ThumbnailResult{}, fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return ThumbnailResult{
		WebPath:       fullDestPath, // Pełna ścieżka dla Wailsa
		Metadata:      metadata,
		IsPlaceholder: false,
	}, nil
}

func (g *ThumbnailGenerator) extractMetadataFromImage(img image.Image) ImageMetadata {
	bounds := img.Bounds()
	domColorHex, err := CalculateDominantColor(img)
	if err != nil {
		g.logger.Warn("Failed to calc dominant color", "err", err)
	}
	closestColor, err := FindClosestPaletteColor(domColorHex, predefinedPalette)
	if err != nil {
		g.logger.Warn("Failed to calc closest color", "err", err)
	}
	meta := ImageMetadata{
		Width:           bounds.Dx(),
		Height:          bounds.Dy(),
		DominantColor:   closestColor,
		BitDepth:        GetBitDepth(img),
		HasAlphaChannel: hasAlpha(img),
	}
	return meta
}

func hasAlpha(img image.Image) bool {
	// Prosta heurystyka bazująca na modelu kolorów
	switch img.ColorModel() {
	case color.RGBAModel, color.NRGBAModel, color.AlphaModel, color.Alpha16Model, color.NYCbCrAModel:
		return true
	}
	return false
}

func (g *ThumbnailGenerator) getPlaceholderResult(ext string) ThumbnailResult {
	return ThumbnailResult{
		WebPath:       g.getPlaceholderPath(ext),
		Metadata:      ImageMetadata{},
		IsPlaceholder: true,
	}
}
func (g *ThumbnailGenerator) getPlaceholderPath(ext string) string {
	const defaultPlaceholder = "generic_placeholder.webp"
	const placeholderPrefix = "/placeholders/"

	fileName, exists := g.placeholderMap[ext]
	if !exists {
		fileName = defaultPlaceholder
	}
	return path.Join(placeholderPrefix, fileName)
}
