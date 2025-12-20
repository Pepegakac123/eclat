package services

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineFileType(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		expected  string
	}{
		{"Standard JPG", ".jpg", "image"},
		{"Case Insensitive PNG", ".PNG", "image"},
		{"Blender File", ".blend", "model"},
		{"Maya ASCII", ".ma", "model"},
		{"Photoshop File", ".psd", "texture"},
		{"Substance Painter", ".spp", "texture"},
		{"Unknown Executable", ".exe", "other"},
		{"No Extension", "filename_without_ext", "other"},
		{"Empty String", "", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineFileType(tt.extension)
			assert.Equal(t, tt.expected, result, "Dla rozszerzenia %s oczekiwano %s", tt.extension, tt.expected)
		})
	}
}

func TestFindClosestPaletteColor(t *testing.T) {
	testPalette := []PaletteColor{
		{Name: "Pure Red", Hex: "#FF0000"},
		{Name: "Pure Green", Hex: "#00FF00"},
		{Name: "Pure Blue", Hex: "#0000FF"},
		{Name: "Black", Hex: "#000000"},
		{Name: "White", Hex: "#FFFFFF"},
	}

	tests := []struct {
		name        string
		inputHex    string
		expectedHex string
		expectError bool
	}{
		{"Exact Match Red", "#FF0000", "#FF0000", false},
		{"Almost Red (Darker)", "#CC0000", "#FF0000", false}, // Powinien przyciągnąć do czerwonego
		{"Almost White (Light Gray)", "#F0F0F0", "#FFFFFF", false},
		{"Invalid Hex", "giberish", "", true},
		{"Short Hex (Supported!)", "#FFF", "#FFFFFF", false}, // Go-colorful wymaga #RRGGBB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindClosestPaletteColor(tt.inputHex, testPalette)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHex, result)
			}
		})
	}
}

// 3. Testujemy analizę obrazu (Trick: tworzymy puste obrazy w pamięci)
func TestGetBitDepth(t *testing.T) {
	t.Run("Should detect 8-bit RGBA", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		depth := GetBitDepth(img)
		assert.Equal(t, 8, depth)
	})

	t.Run("Should detect 16-bit Gray", func(t *testing.T) {
		img := image.NewGray16(image.Rect(0, 0, 10, 10))
		depth := GetBitDepth(img)
		assert.Equal(t, 16, depth)
	})

	t.Run("Should detect 16-bit RGBA64", func(t *testing.T) {
		img := image.NewRGBA64(image.Rect(0, 0, 10, 10))
		depth := GetBitDepth(img)
		assert.Equal(t, 16, depth)
	})

	t.Run("Should handle nil image safe", func(t *testing.T) {
		depth := GetBitDepth(nil)
		assert.Equal(t, 0, depth)
	})
}
