package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"path/filepath"
	"regexp"
	"strings"
)

// versionSuffixes defines regex patterns for common file versioning conventions.
// These are used to strip suffixes like "_v1", " copy", etc., to find the base name.
var versionSuffixes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)[_ -]v\d+$`),          // _v1, -v02, v3
	regexp.MustCompile(`(?i)[_ -]ver\d+$`),        // _ver1
	regexp.MustCompile(`(?i)[_ -]version\s*\d+$`), // _version 1
	regexp.MustCompile(`(?i)[_ -]copy(\s*\d+)?$`), // _copy, -copy 2
	regexp.MustCompile(`(?i)\s*\(\d+\)$`),         // (1), (2) - Windows/Linux system duplicates
	regexp.MustCompile(`(?i)[_ -]final$`),         // _final
}

// getBaseName strips version suffixes from a filename to determine its canonical "base" name.
// It iteratively applies regex patterns until the name stabilizes.
func (s *Scanner) getBaseName(filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))

	for {
		original := base
		for _, re := range versionSuffixes {
			base = re.ReplaceAllString(base, "")
		}
		if base == original {
			break
		}
		base = strings.TrimSpace(base)
	}

	return base
}

// TryHeuristicMatch attempts to find an existing asset group for a file based on its name.
// It strips version suffixes and looks for potential siblings in the same folder.
// Returns the GroupID if a match is found, otherwise empty string.
func (s *Scanner) TryHeuristicMatch(ctx context.Context, folderID int64, filename string) (string, bool) {
	baseName := s.getBaseName(filename)

	if len(baseName) < 3 {
		return "", false
	}

	// Ask DB for candidates in the same folder matching "BaseName%"
	pattern := baseName + "%"
	candidates, err := s.db.FindPotentialSiblings(ctx, database.FindPotentialSiblingsParams{
		ScanFolderID: sql.NullInt64{Int64: folderID, Valid: true},
		FileName:     pattern,
		ID:           0,
	})

	if err != nil {
		s.logger.Warn("Heuristic SQL lookup failed", "error", err)
		return "", false
	}

	// Verify candidates: SQL LIKE is loose, so we verify if candidates
	// actually reduce to the exact same base name.
	for _, cand := range candidates {
		candidateBase := s.getBaseName(cand.FileName)

		if candidateBase == baseName {
			return cand.GroupID, true
		}
	}

	return "", false
}
