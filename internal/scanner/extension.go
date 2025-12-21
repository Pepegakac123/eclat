package scanner

import (
	"eclat/internal/config"
	"fmt"
	"slices"
	"strings"
)

// GetConfig returns a thread-safe copy of configuration
func (s *Scanner) GetConfig() config.ScannerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	safeExts := make([]string, len(s.config.AllowedExtensions))
	copy(safeExts, s.config.AllowedExtensions)

	return config.ScannerConfig{
		AllowedExtensions:    safeExts,
		MaxAllowHashFileSize: s.config.MaxAllowHashFileSize,
	}
}

// GetPredefinedPalette returns color palette
func (s *Scanner) GetPredefinedPalette() []config.PaletteColor {
	return config.PredefinedPalette
}

// IsExtensionAllowed checks if file should be scanned
func (s *Scanner) IsExtensionAllowed(ext string) bool {
	normalized := strings.ToLower(ext)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Contains(s.config.AllowedExtensions, normalized)
}

// AddExtensions safely adds new extensions
func (s *Scanner) AddExtensions(exts []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var invalidExts []string
	if len(exts) <= 0 {
		return nil
	}

	for _, ext := range exts {
		normalized := strings.ToLower(ext)
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}

		if !config.IsExtensionValid(normalized) {
			invalidExts = append(invalidExts, ext)
			continue
		}

		if !slices.Contains(s.config.AllowedExtensions, normalized) {
			s.config.AllowedExtensions = append(s.config.AllowedExtensions, normalized)
		}
	}

	if len(invalidExts) > 0 {
		return fmt.Errorf("invalid or dangerous extensions: %s", strings.Join(invalidExts, ", "))
	}
	return nil
}

// RemoveExtension removes extension from allowed list
func (s *Scanner) RemoveExtension(ext string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(ext)
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}

	s.config.AllowedExtensions = slices.DeleteFunc(s.config.AllowedExtensions, func(e string) bool {
		return e == normalized
	})
}
