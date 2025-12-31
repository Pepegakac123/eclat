package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"fmt"
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
	regexp.MustCompile(`(?i)[_ -]robocze$`),       // _robocze (Polish for working)
	regexp.MustCompile(`(?i)[_ -](work|working|backup|temp|old)$`), // Common variants
	regexp.MustCompile(`(?i)[_ -]\d+$`),           // _001, -02, etc. (trailing numbers with separator)
}

var leadingNumber = regexp.MustCompile(`(?i)^\d+[._ -]+`)

// getBaseName strips version suffixes from a filename to determine its canonical "base" name.
// It iteratively applies regex patterns until the name stabilizes.
func (s *Scanner) getBaseName(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	base = strings.ToLower(base)

	s.logger.Debug("🧠 getBaseName: starting", "file", filename, "ext", ext, "initial_base", base)

	for {
		original := base
		// Strip leading numbers
		base = leadingNumber.ReplaceAllString(base, "")
		if base != original {
			s.logger.Debug("🧠 getBaseName: stripped leading numbers", "from", original, "to", base)
		}

		// Strip version suffixes
		for _, re := range versionSuffixes {
			stripped := re.ReplaceAllString(base, "")
			if stripped != base {
				s.logger.Debug("🧠 getBaseName: stripped suffix", "re", re.String(), "from", base, "to", stripped)
				base = stripped
			}
		}

		if base == original {
			break
		}
		base = strings.TrimSpace(base)
	}

	s.logger.Debug("🧠 getBaseName: final", "file", filename, "base", base)
	return base
}

// TryHeuristicMatch attempts to find an existing asset group for a file based on its name.
// It strips version suffixes and looks for potential siblings in the same folder.
// Returns the GroupID if a match is found, otherwise empty string.
func (s *Scanner) TryHeuristicMatch(ctx context.Context, folderID int64, filename string) (string, bool) {
	baseName := s.getBaseName(filename)

	if len(baseName) < 3 {
		s.logger.Debug("🧠 Heuristic match: base name too short", "file", filename, "base", baseName)
		return "", false
	}

	// 1. Check Session Cache (In-memory grouping for files in the current scan)
	s.sessionMu.Lock()
	cacheKey := fmt.Sprintf("%d:%s", folderID, baseName)
	if cachedGroupID, ok := s.sessionHeuristicCache[cacheKey]; ok {
		s.sessionMu.Unlock()
		s.logger.Debug("🧠 Heuristic match: SUCCESS (Session Cache)", "file", filename, "base", baseName, "group_id", cachedGroupID)
		return cachedGroupID, true
	}
	s.sessionMu.Unlock()

	// 2. Check DB (Previously indexed files)
	// Ask DB for candidates in the same folder matching "%BaseName%"
	// Using % at the beginning allows matching files that might have leading numbers or prefixes 
	// that our getBaseName strips. We then verify the match in Go logic.
	pattern := "%" + baseName + "%"
	s.logger.Debug("🧠 Heuristic match: looking for siblings in DB", "file", filename, "base", baseName, "pattern", pattern)

	candidates, err := s.db.FindPotentialSiblings(ctx, database.FindPotentialSiblingsParams{
		ScanFolderID: sql.NullInt64{Int64: folderID, Valid: true},
		FileName:     pattern,
		ID:           0,
		Limit:        50,
	})

	if err != nil {
		s.logger.Warn("Heuristic SQL lookup failed", "error", err)
		return "", false
	}

	s.logger.Debug("🧠 Heuristic match: DB returned candidates", "count", len(candidates))

	// Verify candidates: SQL LIKE is loose, so we verify if candidates
	// actually reduce to the exact same base name.
	for _, cand := range candidates {
		candidateBase := s.getBaseName(cand.FileName)
		s.logger.Debug("🧠 Heuristic match: checking candidate", "file", cand.FileName, "candidate_base", candidateBase, "target_base", baseName)

		if candidateBase == baseName {
			s.logger.Debug("🧠 Heuristic match: SUCCESS", "file", filename, "matched_with", cand.FileName, "group_id", cand.GroupID)
			return cand.GroupID, true
		}
	}

	s.logger.Debug("🧠 Heuristic match: NO MATCH found", "file", filename, "base", baseName)
	return "", false
}
