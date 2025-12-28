package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"path/filepath"
	"regexp"
	"strings"
)

var versionSuffixes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)[_ -]v\d+$`),          // _v1, -v02, v3
	regexp.MustCompile(`(?i)[_ -]ver\d+$`),        // _ver1
	regexp.MustCompile(`(?i)[_ -]version\s*\d+$`), // _version 1
	regexp.MustCompile(`(?i)[_ -]copy(\s*\d+)?$`), // _copy, -copy 2
	regexp.MustCompile(`(?i)\s*\(\d+\)$`),         // (1), (2) - systemowe duplikaty Windows/Linux
	regexp.MustCompile(`(?i)[_ -]final$`),         // _final
}

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

// TryHeuristicMatch to nasz detektyw. Próbuje znaleźć grupę dla samotnego pliku.
func (s *Scanner) TryHeuristicMatch(ctx context.Context, folderID int64, filename string) (string, bool) {
	baseName := s.getBaseName(filename)

	if len(baseName) < 3 {
		return "", false
	}

	// Zapytaj bazę o kandydatów w tym samym folderze
	// Szukamy po "Rdzeń%", np. "Monster%"
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

	// 3. Weryfikacja kandydatów
	// SQL LIKE jest "luźny", więc musimy sprawdzić, czy kandydaci
	// faktycznie redukują się do tego samego rdzenia.
	for _, cand := range candidates {
		candidateBase := s.getBaseName(cand.FileName)

		if candidateBase == baseName {
			return cand.GroupID, true
		}
	}

	return "", false
}
