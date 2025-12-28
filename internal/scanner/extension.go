package scanner

import (
	"context"
	"eclat/internal/config"
	"eclat/internal/database"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// ScannerConfigSnapshot to struktura publiczna (DTO) dla Frontendu.
type ScannerConfigSnapshot struct {
	AllowedExtensions    []string `json:"allowedExtensions"`
	MaxAllowHashFileSize int64    `json:"maxAllowHashFileSize"`
}

// GetConfig returns a thread-safe snapshot of configuration for the UI
func (s *Scanner) GetConfig() ScannerConfigSnapshot {
	return ScannerConfigSnapshot{
		AllowedExtensions:    s.config.GetAllowedExtensions(),
		MaxAllowHashFileSize: s.config.GetMaxHashFileSize(),
	}
}

// GetPredefinedPalette returns color palette
func (s *Scanner) GetPredefinedPalette() []config.PaletteColor {
	return config.PredefinedPalette
}

// IsExtensionAllowed checks if file should be scanned
func (s *Scanner) IsExtensionAllowed(ext string) bool {
	// Delegujemy do configa - on ma RLocka w środku.
	return s.config.IsExtensionAllowed(ext)
}

// AddExtensions safely adds new extensions
func (s *Scanner) AddExtensions(exts []string) error {
	// 1. Walidacja (statyczna, nie wymaga locka)
	var invalidExts []string
	if len(exts) <= 0 {
		return nil
	}

	for _, ext := range exts {
		if !config.IsExtensionValid(ext) {
			invalidExts = append(invalidExts, ext)
		}
	}

	if len(invalidExts) > 0 {
		return fmt.Errorf("invalid or dangerous extensions: %s", strings.Join(invalidExts, ", "))
	}

	// 2. Pobieramy obecny stan (Thread-Safe Copy)
	currentExts := s.config.GetAllowedExtensions()
	modified := false

	// 3. Modyfikujemy lokalną kopię
	for _, ext := range exts {
		normalized := strings.ToLower(ext)
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}

		if !slices.Contains(currentExts, normalized) {
			currentExts = append(currentExts, normalized)
			modified = true
		}
	}

	if modified {
		s.config.SetAllowedExtensions(currentExts)
		// PERSISTENCE: Zapisujemy do bazy
		if err := s.persistExtensions(currentExts); err != nil {
			s.logger.Error("Failed to persist extensions", "error", err)
			return err
		}
	}

	return nil
}

// RemoveExtension removes extension from allowed list
func (s *Scanner) RemoveExtension(ext string) {
	normalized := strings.ToLower(ext)
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}

	currentExts := s.config.GetAllowedExtensions()

	newExts := slices.DeleteFunc(currentExts, func(e string) bool {
		return e == normalized
	})

	// Jeśli długość się zmieniła, aktualizujemy
	if len(newExts) != len(currentExts) {
		s.config.SetAllowedExtensions(newExts)
		// PERSISTENCE: Zapisujemy do bazy (ignorujemy błąd w sygnaturze, ale logujemy)
		if err := s.persistExtensions(newExts); err != nil {
			s.logger.Error("Failed to persist extensions removal", "error", err)
		}
	}
}

// persistExtensions to helper do zapisywania stanu w bazie
func (s *Scanner) persistExtensions(exts []string) error {
	// Safety check dla testów, gdzie db może być nil
	if s.db == nil {
		return nil
	}

	jsonBytes, err := json.Marshal(exts)
	if err != nil {
		return err
	}

	// Używamy kontekstu aplikacji lub tła, jeśli jeszcze nie wystartowała
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return s.db.SetSystemSetting(ctx, database.SetSystemSettingParams{
		Key:   "allowed_extensions",
		Value: string(jsonBytes),
	})
}
