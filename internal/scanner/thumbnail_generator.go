package scanner

import (
	"context"
	"eclat/internal/config"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// ThumbnailGenerator defines the interface for generating or retrieving thumbnails for assets.
type ThumbnailGenerator interface {
	Generate(ctx context.Context, sourcePath string) (ThumbnailResult, error)
}

// DiskThumbnailGenerator implements ThumbnailGenerator by creating thumbnails on disk
// or returning static placeholders for unsupported file types.
type DiskThumbnailGenerator struct {
	cacheDir       string
	logger         *slog.Logger
	placeholderMap map[string]string
}

// ThumbnailResult contains the path to the generated thumbnail (or placeholder)
// and extracted image metadata.
type ThumbnailResult struct {
	WebPath       string
	Metadata      ImageMetadata
	IsPlaceholder bool
}

// NewDiskThumbnailGenerator creates a new generator that stores thumbnails in the specified cache directory.
// It initializes a map of default placeholders for various 3D and texture formats.
func NewDiskThumbnailGenerator(cacheDir string, logger *slog.Logger) *DiskThumbnailGenerator {
	return &DiskThumbnailGenerator{
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

// Generate creates a thumbnail for the file at srcPath.
// If the file is a supported image, it generates a WebP thumbnail and extracts metadata.
// For other files, it returns a pre-configured placeholder.
func (g *DiskThumbnailGenerator) Generate(ctx context.Context, srcPath string) (ThumbnailResult, error) {
	ext := strings.ToLower(filepath.Ext(srcPath))

	// 1. If it's a supported image format, try to generate a real thumbnail
	if isSupportedImageExt(ext) {
		g.logger.Debug("Generating thumbnail from image", "path", srcPath)
		res, err := g.generateFromImage(srcPath)
		if err == nil {
			return res, nil
		}
		// If generation failed (e.g. corrupt file), log and fall back to placeholder
		g.logger.Warn("Thumbnail generation failed, falling back to placeholder", "path", srcPath, "error", err)
	}

	// 2. Fallback to placeholder for non-image types or failed generation
	res := g.getPlaceholderResult(ext)
	g.logger.Debug("Using placeholder", "path", srcPath, "placeholder", res.WebPath)
	return res, nil
}

func (g *DiskThumbnailGenerator) generateFromImage(srcPath string) (ThumbnailResult, error) {
	img, err := imaging.Open(srcPath)
	if err != nil {
		return ThumbnailResult{}, fmt.Errorf("failed to open image: %w", err)
	}
	originalBounds := img.Bounds()
	hasAlphaChannel := hasAlpha(img)
	bitDepth := GetBitDepth(img)
	thumb := imaging.Resize(img, 400, 0, imaging.Linear)

	imgMetadata := g.extractMetadataFromThumb(thumb, originalBounds, bitDepth, hasAlphaChannel)

	bounds := thumb.Bounds()
	imgRGBA := image.NewRGBA(bounds)
	draw.Draw(imgRGBA, bounds, thumb, bounds.Min, draw.Src)

	id := uuid.New()
	filename := fmt.Sprintf("%s.webp", id.String())
	fullDestPath := filepath.Join(g.cacheDir, filename)

	outFile, err := os.Create(fullDestPath)
	if err != nil {
		return ThumbnailResult{}, fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer outFile.Close()

	err = webp.Encode(outFile, imgRGBA, &webp.Options{
		Lossless: false,
		Quality:  80,
	})
	if err != nil {
		return ThumbnailResult{}, fmt.Errorf("webp encode error: %w", err)
	}

	return ThumbnailResult{
		WebPath:       "/thumbnails/" + filename,
		Metadata:      imgMetadata,
		IsPlaceholder: false,
	}, nil
}

func (g *DiskThumbnailGenerator) extractMetadataFromThumb(thumb image.Image, origBounds image.Rectangle, bitDepth int, hasAlpha bool) ImageMetadata {

	domColorHex, err := CalculateDominantColor(thumb)

	var closestColor string
	var hexColor string

	if err != nil {

		g.logger.Debug("Skipping color analysis", "reason", err)
		hexColor = ""
	} else {
		hexColor = domColorHex
		closest, err := FindClosestPaletteColor(domColorHex, config.PredefinedPalette)
		if err != nil {
			g.logger.Debug("Failed to find closest color", "hex", hexColor, "error", err)
		} else {
			closestColor = closest
		}
	}

	meta := ImageMetadata{
		Width:           origBounds.Dx(),
		Height:          origBounds.Dy(),
		DominantColor:   closestColor,
		BitDepth:        bitDepth,
		HasAlphaChannel: hasAlpha,
	}
	return meta
}

func hasAlpha(img image.Image) bool {
	switch img.ColorModel() {
	case color.RGBAModel, color.NRGBAModel, color.AlphaModel, color.Alpha16Model, color.NYCbCrAModel:
		return true
	}
	return false
}

func (g *DiskThumbnailGenerator) getPlaceholderResult(ext string) ThumbnailResult {
	return ThumbnailResult{
		WebPath:       g.getPlaceholderPath(ext),
		Metadata:      ImageMetadata{},
		IsPlaceholder: true,
	}
}

func (g *DiskThumbnailGenerator) getPlaceholderPath(ext string) string {
	const defaultPlaceholder = "generic_placeholder.webp"
	const placeholderPrefix = "/placeholders/"

	fileName, exists := g.placeholderMap[ext]
	if !exists {
		fileName = defaultPlaceholder
	}
	return path.Join(placeholderPrefix, fileName)
}
