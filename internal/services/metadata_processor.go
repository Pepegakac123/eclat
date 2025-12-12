package services

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/lucasb-eyer/go-colorful"
)

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
