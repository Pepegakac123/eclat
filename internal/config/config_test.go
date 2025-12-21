package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExtensionValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid JPG", "image.jpg", true},
		{"Valid JPG Uppercase", "IMAGE.JPG", true}, // Case insensitivity
		{"Valid Blend", "model.blend", true},

		{"Dangerous EXE", "virus.exe", false},
		{"Dangerous EXE Uppercase", "VIRUS.EXE", false},
		{"Dangerous BAT", "script.bat", false},

		{"No dot dangerous", "exe", false}, // Funkcja dodaje kropkę, więc exe -> .exe -> block
		{"No dot valid", "jpg", true},      // Funkcja dodaje kropkę, więc jpg -> .jpg -> pass

		{"Empty string", "", false},
		{"Just dot", ".", false},
		{"Dot at end", "image.", false}, // Ext zwraca puste lub kropkę
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExtensionValid(tt.input)
			assert.Equal(t, tt.want, got, "Dla wejścia: %s", tt.input)
		})
	}
}

func TestNewScannerConfig_Defaults(t *testing.T) {
	cfg := NewScannerConfig()

	assert.NotNil(t, cfg)
	assert.Greater(t, len(cfg.AllowedExtensions), 0, "Lista rozszerzeń nie powinna być pusta")
	assert.Equal(t, int64(268435456), cfg.MaxAllowHashFileSize, "Domyślny limit rozmiaru pliku nieprawidłowy")
}

func TestNewScannerConfig_SliceIsolation(t *testing.T) {
	// Ten test chroni przed usunięciem 'copy()' z konstruktora.
	// W Go slice to referencja. Bez copy(), modyfikacja cfg zmieniłaby DefaultAllowedExtensions.

	cfg := NewScannerConfig()

	// Zapamiętujemy oryginał
	originalFirst := DefaultAllowedExtensions[0]

	// Modyfikujemy instancję
	cfg.AllowedExtensions[0] = ".HACKED"

	// Sprawdzamy czy globalna zmienna pozostała nienaruszona
	assert.Equal(t, originalFirst, DefaultAllowedExtensions[0],
		"CRITICAL: Modyfikacja instancji configu nadpisała globalną zmienną! Brakuje copy() w konstruktorze?")

	assert.NotEqual(t, cfg.AllowedExtensions[0], DefaultAllowedExtensions[0])
}
