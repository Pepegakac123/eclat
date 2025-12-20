package services

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskThumbnailGenerator_Placeholders(t *testing.T) {
	// 1. Setup
	// Nie potrzebujemy prawdziwych plików na dysku do testowania placeholderów,
	// bo logika sprawdza tylko rozszerzenie stringa.
	cacheDir := "/tmp/cache" // Ścieżka nie ma znaczenia w tym teście
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Tworzymy PRAWDZIWĄ implementację (nie Mocka)
	generator := NewDiskThumbnailGenerator(cacheDir, logger)

	tests := []struct {
		name              string
		filename          string
		expectedIcon      string
		expectPlaceholder bool
	}{
		{
			name:              "Blender File",
			filename:          "project.blend",
			expectedIcon:      "blend_placeholder.webp",
			expectPlaceholder: true,
		},
		{
			name:              "Maya File",
			filename:          "model.ma",
			expectedIcon:      "ma_placeholder.webp",
			expectPlaceholder: true,
		},
		{
			name:              "Substance Painter",
			filename:          "texture.spp",
			expectedIcon:      "spp_placeholder.webp",
			expectPlaceholder: true,
		},
		{
			name:              "Unknown File Extension",
			filename:          "document.pdf",
			expectedIcon:      "generic_placeholder.webp", // Domyślny fallback
			expectPlaceholder: true,
		},
		{
			name:              "File without extension",
			filename:          "README",
			expectedIcon:      "generic_placeholder.webp",
			expectPlaceholder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ACTION
			// Generate dla plików, które nie są obrazkami (np. .blend),
			// powinno od razu zwrócić placeholder bez dotykania dysku.
			result, err := generator.Generate(context.Background(), tt.filename)

			// ASSERT
			assert.NoError(t, err)
			assert.Equal(t, tt.expectPlaceholder, result.IsPlaceholder)
			assert.Contains(t, result.WebPath, tt.expectedIcon, "Ścieżka powinna zawierać odpowiednią ikonę")
			assert.Contains(t, result.WebPath, "/placeholders/", "Ścieżka powinna wskazywać na folder placeholders")
		})
	}
}

// Ten test sprawdza czy struktura implementuje interfejs (compile-time check)
func TestDiskThumbnailGenerator_ImplementsInterface(t *testing.T) {
	var _ ThumbnailGenerator = (*DiskThumbnailGenerator)(nil)
}
