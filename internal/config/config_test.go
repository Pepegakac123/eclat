package config

import (
	"sync"
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

	allowed := cfg.GetAllowedExtensions()
	assert.Greater(t, len(allowed), 0, "Lista rozszerzeń nie powinna być pusta")

	assert.Equal(t, int64(268435456), cfg.GetMaxHashFileSize(), "Domyślny limit rozmiaru pliku nieprawidłowy")
}

func TestNewScannerConfig_SliceIsolation(t *testing.T) {
	cfg := NewScannerConfig()

	originalFirst := DefaultAllowedExtensions[0]
	currentExts := cfg.GetAllowedExtensions()

	currentExts[0] = ".HACKED"
	cfg.SetAllowedExtensions([]string{".NEW", ".STUFF"})
	assert.Equal(t, originalFirst, DefaultAllowedExtensions[0],
		"CRITICAL: Globalna zmienna została naruszona!")

	newConfigExts := cfg.GetAllowedExtensions()
	assert.Equal(t, ".NEW", newConfigExts[0])
	assert.NotContains(t, newConfigExts, ".HACKED", "Config nie powinien przyjąć modyfikacji lokalnej kopii")
}

func TestScannerConfig_Concurrency(t *testing.T) {
	cfg := NewScannerConfig()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cfg.IsExtensionAllowed("test.jpg")
			_ = cfg.GetAllowedExtensions()
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg.SetAllowedExtensions([]string{".png", ".jpg"})
		}()
	}

	wg.Wait()
	assert.True(t, cfg.IsExtensionAllowed("test.jpg"))
}
