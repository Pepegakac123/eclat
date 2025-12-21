package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 1. TEST: ZOMBIE RESURRECTION 🧟
// Sprawdza czy usunięty (soft-delete) plik wraca do życia, gdy pojawi się na dysku.
func TestScanner_Logic_Resurrection(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	filename := "zombie.txt"
	path := filepath.Join(root, filename)

	// A. Wrzucamy do bazy asset oznaczony jako USUNIĘTY
	asset := insertTestAsset(t, queries, 1, path, "dummy_hash")
	queries.SoftDeleteAsset(ctx, asset.ID)

	// B. Tworzymy plik fizycznie na dysku (Użytkownik go przywrócił)
	createDummyFile(t, path)

	// C. Skanujemy
	err := scanner.StartScan()
	assert.NoError(t, err)

	// Czekamy na wynik
	assert.Eventually(t, func() bool {
		a, err := queries.GetAssetById(ctx, asset.ID)
		return err == nil && !a.IsDeleted // Musi być !IsDeleted
	}, 2*time.Second, 50*time.Millisecond, "Zombie asset powinien zostać przywrócony")
}

// 2. TEST: MOVE DETECTION (RENAME) 🚚
// Sprawdza czy zmiana nazwy pliku aktualizuje ścieżkę w bazie (zachowując ID), a nie tworzy duplikatu.
func TestScanner_Logic_MoveDetection(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	oldPath := filepath.Join(root, "old_name.txt")
	newPath := filepath.Join(root, "new_name.txt")

	createDummyFile(t, newPath)

	realHash, _ := CalculateFileHash(newPath, 0)

	oldAsset := insertTestAsset(t, queries, 1, oldPath, realHash)

	err := scanner.StartScan()
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		// 1. Stary asset powinien mieć zaktualizowaną ścieżkę
		updatedAsset, err := queries.GetAssetById(ctx, oldAsset.ID)
		if err != nil {
			return false
		}

		// ID musi być to samo, Path musi być nowa
		return updatedAsset.FilePath == newPath && !updatedAsset.IsDeleted
	}, 2*time.Second, 50*time.Millisecond, "Asset powinien zostać przeniesiony (zaktualizowana ścieżka)")

	// Sprawdźmy czy nie ma duplikatu (czy stary nie został oznaczony jako deleted, a nowy dodany)
	assets, _ := queries.ListAssets(ctx, database.ListAssetsParams{Limit: 10})

	// Powinien być TYLKO 1 asset (ten przeniesiony)
	// (Chyba że soft delete zostawia rekordy, wtedy sprawdzamy czy nie ma nowego ID)
	assert.Equal(t, 1, len(assets), "Nie powinno być duplikatów po zmianie nazwy")
	assert.Equal(t, oldAsset.ID, assets[0].ID, "ID powinno zostać zachowane")
}

// 3. TEST: COPY DETECTION (DUPLICATE) 👯
// Sprawdza czy skopiowanie pliku tworzy nowy wpis w bazie.
func TestScanner_Logic_CopyDetection(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	path1 := filepath.Join(root, "original.txt")
	path2 := filepath.Join(root, "copy.txt")

	// Tworzymy oba pliki na dysku (ta sama treść = ten sam hash)
	createDummyFile(t, path1)
	createDummyFile(t, path2)

	realHash, _ := CalculateFileHash(path1, 0)

	// A. W bazie jest tylko oryginał
	insertTestAsset(t, queries, 1, path1, realHash)

	// B. Skanujemy
	err := scanner.StartScan()
	assert.NoError(t, err)

	// C. Weryfikacja
	assert.Eventually(t, func() bool {
		assets, _ := queries.ListAssets(ctx, database.ListAssetsParams{Limit: 10})
		return len(assets) == 2
	}, 2*time.Second, 50*time.Millisecond, "Powinny być 2 assety (oryginał i kopia)")
}

// 4. TEST: CLEANUP (SOFT DELETE) 🗑️
// Sprawdza czy plik usunięty z dysku dostaje flagę IsDeleted w bazie.
func TestScanner_Logic_Cleanup(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	missingPath := filepath.Join(root, "missing.txt")
	existingPath := filepath.Join(root, "existing.txt")

	// A. Na dysku tworzymy tylko "existing"
	createDummyFile(t, existingPath)

	// B. W bazie mamy oba (jeden to "duch")
	insertTestAsset(t, queries, 1, missingPath, "some_hash")
	insertTestAsset(t, queries, 1, existingPath, "some_hash_2")

	// C. Skanujemy
	err := scanner.StartScan()
	assert.NoError(t, err)

	// D. Weryfikacja
	assert.Eventually(t, func() bool {
		// Existing powinien być active
		a1, _ := queries.GetAssetByPath(ctx, existingPath)
		if a1.IsDeleted {
			return false
		}

		// Missing powinien być soft deleted
		a2, _ := queries.GetAssetByPath(ctx, missingPath)
		return a2.IsDeleted
	}, 2*time.Second, 50*time.Millisecond, "Brakujący plik powinien dostać Soft Delete")
}

// 5. TEST: METADATA REFRESH 📝
// Sprawdza czy zmiana zawartości pliku wymusza odświeżenie metadanych.
func TestScanner_Logic_MetadataRefresh(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	path := filepath.Join(root, "data.txt")
	createDummyFile(t, path) // Rozmiar X, Czas Y

	// A. W bazie mamy stare dane (symulujemy że plik był mniejszy i starszy)
	asset := insertTestAsset(t, queries, 1, path, "old_hash")

	// Ręcznie psujemy metadane w bazie, żeby zobaczyć czy się naprawią
	// Zmieniamy LastModified na bardzo stary
	oldTime := time.Now().Add(-24 * time.Hour)
	err := queries.RefreshAssetTechnicalMetadata(ctx, database.RefreshAssetTechnicalMetadataParams{
		FileSize:        0,
		LastModified:    oldTime,
		LastScanned:     oldTime,
		ThumbnailPath:   "",
		ImageWidth:      sql.NullInt64{Valid: false},
		ImageHeight:     sql.NullInt64{Valid: false},
		DominantColor:   sql.NullString{Valid: false},
		BitDepth:        sql.NullInt64{Valid: false},
		HasAlphaChannel: sql.NullBool{Valid: false},

		ID: asset.ID,
	})
	assert.NoError(t, err)

	// B. Skanujemy
	err = scanner.StartScan()
	assert.NoError(t, err)

	// C. Weryfikacja
	assert.Eventually(t, func() bool {
		updated, err := queries.GetAssetById(ctx, asset.ID)
		if err != nil {
			return false
		}

		// Sprawdzamy czy LastModified się zaktualizował (powinien być bliski teraz, a nie wczoraj)
		return updated.FileSize > 0 && updated.LastModified.After(oldTime)
	}, 2*time.Second, 50*time.Millisecond, "Metadane powinny zostać odświeżone")
}
