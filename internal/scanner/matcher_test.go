package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUnit_GetBaseName sprawdza tylko i wyłącznie logikę Regexów.
// Nie dotyka bazy danych, dysku ani reszty systemu.
func TestUnit_GetBaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Proste przypadki
		{"Monster.png", "monster"},
		{"Monster_v1.png", "monster"},
		{"Monster-v2.jpg", "monster"},

		// Złożone wersjonowanie
		{"Hero_Character_final.obj", "hero_character"},
		{"Hero_Character_FINAL_v3.obj", "hero_character"}, // Wielokrotne czyszczenie
		{"Weapon Sword copy.fbx", "weapon sword"},
		{"Weapon Sword copy 2.fbx", "weapon sword"},

		// Systemowe duplikaty
		{"Texture (1).png", "texture"},
		{"Texture (2).jpg", "texture"},

		// Blender backups
		{"ork.blend1", "ork"},
		{"ork.blend2", "ork"},

		// Numeric suffixes with separators
		{"GNOM_001.zpr", "gnom"},
		{"GNOM_002.zpr", "gnom"},
		{"Asset_01.obj", "asset"},
		{"Asset-02.obj", "asset"},

		// Polish/Common working file patterns
		{"ork_robocze_01.blend", "ork"},
		{"model_final_v2.obj", "model"},

		// Leading numbers
		{"7. Character-Sheet-Orc.jpg", "character-sheet-orc"},
		{"01_Background.png", "background"},

		// Edge cases - nazwy, których nie powinno zepsuć
		{"version_control.txt", "version_control"}, // Nie jest na końcu
		{"my_vacation.jpg", "my_vacation"},         // 'v' w środku słowa
		{"v2.png", "v2"},                           // Zbyt krótka nazwa po wycięciu (zostawi jak jest lub wytnie do zera - zależnie od logiki, tu zakładamy że regex zadziała)
	}

	// Ponieważ getBaseName jest metodą Scannera, potrzebujemy pustej instancji
	s := &Scanner{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := s.getBaseName(tt.input)
			assert.Equal(t, tt.expected, result, "Dla wejścia '%s' oczekiwano '%s'", tt.input, tt.expected)
		})
	}
}

// TestUnit_TryHeuristicMatch sprawdza, czy funkcja potrafi znaleźć "kolegę" w bazie.
// Wymaga bazy danych (setupLogicTest), ale testuje tylko jedną funkcję izolowaną.
func TestUnit_TryHeuristicMatch(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	// 1. Przygotowanie danych w bazie
	// Wrzucamy plik "Rodzic", który ma już swoje GroupID
	parentName := "Environment_Forest.psd"
	parentPath := filepath.Join(root, parentName)

	// Tworzymy asset ręcznie w bazie (symulacja, że już tam jest)
	_, err := queries.CreateAsset(ctx, database.CreateAssetParams{
		ScanFolderID: sql.NullInt64{Int64: 1, Valid: true}, // Zakładamy folder ID=1 z setupu
		GroupID:      "UUID-RODZINA-LASU",                  // Ustalamy sztywne ID grupy
		FileName:     parentName,
		FilePath:     parentPath,
		LastModified: time.Now(),
		LastScanned:  time.Now(),
	})
	assert.NoError(t, err)

	// Dodajemy też coś z numerkiem na początku
	queries.CreateAsset(ctx, database.CreateAssetParams{
		ScanFolderID: sql.NullInt64{Int64: 1, Valid: true},
		GroupID:      "UUID-GNOM",
		FileName:     "01_Gnom_High.obj",
		FilePath:     filepath.Join(root, "01_Gnom_High.obj"),
		LastModified: time.Now(),
		LastScanned:  time.Now(),
	})

	// 2. Testy Scenariuszowe
	tests := []struct {
		name          string
		filename      string
		shouldMatch   bool
		expectGroupID string
	}{
		{
			name:          "Should match explicit version v2",
			filename:      "Environment_Forest_v2.psd",
			shouldMatch:   true,
			expectGroupID: "UUID-RODZINA-LASU",
		},
		{
			name:          "Should match copy",
			filename:      "Environment_Forest copy.psd",
			shouldMatch:   true,
			expectGroupID: "UUID-RODZINA-LASU",
		},
		{
			name:          "Should match with leading numbers (if already in DB)",
			filename:      "02_Environment_Forest.psd",
			shouldMatch:   true,
			expectGroupID: "UUID-RODZINA-LASU",
		},
		{
			name:          "Should match even if DB entry has leading numbers",
			filename:      "Environment_Forest_Working.psd",
			shouldMatch:   true,
			expectGroupID: "UUID-RODZINA-LASU",
		},
		{
			name:          "Should match Gnom variant",
			filename:      "Gnom_High_v2.obj",
			shouldMatch:   true,
			expectGroupID: "UUID-GNOM",
		},
		{
			name:          "Should NOT match totally different file",
			filename:      "Environment_Desert.psd",
			shouldMatch:   false,
			expectGroupID: "",
		},
		{
			name:          "Should NOT match similar prefix but different base",
			filename:      "Environment_Forest_Fire.psd", // _Fire to nie suffix wersji
			shouldMatch:   false,
			expectGroupID: "", // Tu getBaseName zwróci "Environment_Forest_Fire", co nie pasuje do "Environment_Forest"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupID, found := scanner.TryHeuristicMatch(ctx, 1, tt.filename) // Folder ID = 1

			assert.Equal(t, tt.shouldMatch, found)
			if tt.shouldMatch {
				assert.Equal(t, tt.expectGroupID, groupID)
			}
		})
	}
}
