package scanner

import (
	"eclat/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner_ExtensionLogic(t *testing.T) {
	cfg := config.NewScannerConfig()

	// Stan początkowy
	cfg.SetAllowedExtensions([]string{".jpg", ".png"})

	scanner := &Scanner{
		config: cfg,
	}

	t.Run("Should allow valid extension", func(t *testing.T) {
		result := scanner.IsExtensionAllowed(".jpg")
		assert.True(t, result, "JPG powinno być dozwolone")

		result = scanner.IsExtensionAllowed(".PNG") // Case insensitive check
		assert.True(t, result, "PNG (caps) powinno być dozwolone")
	})

	t.Run("Should reject unknown extension", func(t *testing.T) {
		result := scanner.IsExtensionAllowed(".blend")
		assert.False(t, result, "Blend nie został jeszcze dodany")
	})

	t.Run("AddExtensions - Happy Path", func(t *testing.T) {
		newExts := []string{"blend", ".Obj", "FBX"}

		err := scanner.AddExtensions(newExts)
		assert.NoError(t, err)
		assert.True(t, scanner.IsExtensionAllowed(".blend"))
		assert.True(t, scanner.IsExtensionAllowed(".obj"))
		assert.True(t, scanner.IsExtensionAllowed(".fbx"))
	})

	t.Run("AddExtensions - Security Check", func(t *testing.T) {
		dangerous := []string{".exe", "bat", ".sh"}

		err := scanner.AddExtensions(dangerous)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or dangerous")

		assert.False(t, scanner.IsExtensionAllowed(".exe"))
		assert.False(t, scanner.IsExtensionAllowed(".bat"))
	})

	t.Run("AddExtensions - Deduplication", func(t *testing.T) {
		// Pobieramy stan przez publiczne API
		initialCount := len(scanner.GetConfig().AllowedExtensions)

		err := scanner.AddExtensions([]string{".jpg", "PNG"})
		assert.NoError(t, err)

		finalCount := len(scanner.GetConfig().AllowedExtensions)
		assert.Equal(t, initialCount, finalCount, "Nie powinno dodać duplikatów")
	})

	t.Run("RemoveExtension", func(t *testing.T) {
		// Reset stanu przed testem usuwania
		cfg.SetAllowedExtensions([]string{".jpg", ".png"})

		scanner.RemoveExtension(".jpg")
		assert.False(t, scanner.IsExtensionAllowed(".jpg"))

		scanner.RemoveExtension("png") // Test usuwania bez kropki
		assert.False(t, scanner.IsExtensionAllowed(".png"))
	})

	t.Run("GetConfig returns Snapshot", func(t *testing.T) {
		// NAPRAWA: Musimy upewnić się, że mamy dane testowe, bo poprzedni test (RemoveExtension) wyczyścił wszystko!
		cfg.SetAllowedExtensions([]string{".png", ".jpg"})

		snapshot := scanner.GetConfig()

		// Sprawdzamy czy snapshot zawiera dane
		assert.Contains(t, snapshot.AllowedExtensions, ".png")

		// Sprawdzamy czy modyfikacja snapshota nie psuje oryginału (izolacja)
		// Upewniamy się, że tablica nie jest pusta przed indeksem (dla bezpieczeństwa testu)
		if assert.NotEmpty(t, snapshot.AllowedExtensions) {
			snapshot.AllowedExtensions[0] = ".HACKED"
			assert.False(t, scanner.IsExtensionAllowed(".HACKED"), "Modyfikacja DTO nie powinna wpływać na Scanner")
		}
	})
}
