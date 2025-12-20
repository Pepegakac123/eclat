package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner_ExtensionLogic(t *testing.T) {
	scanner := &Scanner{
		config: &ScannerConfig{
			AllowedExtensions: []string{".jpg", ".png"},
		},
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
		initialCount := len(scanner.config.AllowedExtensions)

		err := scanner.AddExtensions([]string{".jpg", "PNG"})
		assert.NoError(t, err)

		assert.Equal(t, initialCount, len(scanner.config.AllowedExtensions), "Nie powinno dodać duplikatów")
	})

	t.Run("RemoveExtension", func(t *testing.T) {

		scanner.RemoveExtension(".jpg")
		assert.False(t, scanner.IsExtensionAllowed(".jpg"))
		scanner.RemoveExtension("png")
		assert.False(t, scanner.IsExtensionAllowed(".png"))
	})
}
