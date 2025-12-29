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

// ScannerConfigSnapshot is a Data Transfer Object (DTO) representing the current configuration state.
// It is used to pass configuration data to the frontend.
type ScannerConfigSnapshot struct {
	AllowedExtensions    []string `json:"allowedExtensions"`
	MaxAllowHashFileSize int64    `json:"maxAllowHashFileSize"`
}

// GetConfig returns a thread-safe snapshot of the current scanner configuration.
func (s *Scanner) GetConfig() ScannerConfigSnapshot {
	return ScannerConfigSnapshot{
		AllowedExtensions:    s.config.GetAllowedExtensions(),
		MaxAllowHashFileSize: s.config.GetMaxHashFileSize(),
	}
}

// GetPredefinedPalette returns the list of predefined colors used for palette matching.
func (s *Scanner) GetPredefinedPalette() []config.PaletteColor {
	return config.PredefinedPalette
}

// IsExtensionAllowed checks if the given file extension is currently allowed by the configuration.
func (s *Scanner) IsExtensionAllowed(ext string) bool {
	// Delegates to the thread-safe config method.
	return s.config.IsExtensionAllowed(ext)
}

// AddExtensions safely adds a list of new file extensions to the allowed list.
// It performs validation to ensure no dangerous extensions are added and persists the changes to the database.
func (s *Scanner) AddExtensions(exts []string) error {
	// 1. Validation (static check, no lock needed)
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

	// 2. Get current state (Thread-Safe Copy)
	currentExts := s.config.GetAllowedExtensions()
	modified := false

	// 3. Modify local copy
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
		// PERSISTENCE: Save to DB
		if err := s.persistExtensions(currentExts); err != nil {
			s.logger.Error("Failed to persist extensions", "error", err)
			return err
		}
	}

	return nil
}

// RemoveExtension removes a specific extension from the allowed list.
// If the extension was present, the updated list is persisted to the database.
func (s *Scanner) RemoveExtension(ext string) {
	normalized := strings.ToLower(ext)
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}

	currentExts := s.config.GetAllowedExtensions()

	newExts := slices.DeleteFunc(currentExts, func(e string) bool {
		return e == normalized
	})

	// If length changed, update configuration
	if len(newExts) != len(currentExts) {
		s.config.SetAllowedExtensions(newExts)
		// PERSISTENCE: Save to DB (ignoring error in signature, but logging it)
		if err := s.persistExtensions(newExts); err != nil {
			s.logger.Error("Failed to persist extensions removal", "error", err)
		}
	}
}

// persistExtensions helper function to save the list of allowed extensions to the database.
func (s *Scanner) persistExtensions(exts []string) error {
	// Safety check for tests where db might be nil
	if s.db == nil {
		return nil
	}

	jsonBytes, err := json.Marshal(exts)
	if err != nil {
		return err
	}

	// Use application context or background if not started
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return s.db.SetSystemSetting(ctx, database.SetSystemSettingParams{
		Key:   "allowed_extensions",
		Value: string(jsonBytes),
	})
}
