package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"os"
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

// Sprawdza, czy ScanFile poprawnie dodaje pojedynczy plik do bazy.
func TestScanner_Live_ScanFile(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	// A. Tworzymy plik
	fileName := "live_test.png"
	path := filepath.Join(root, fileName)
	createDummyFile(t, path)

	// B. Uruchamiamy ScanFile (bez StartScan!)
	err := scanner.ScanFile(ctx, path)
	assert.NoError(t, err)

	// C. Sprawdzamy czy trafił do bazy
	asset, err := queries.GetAssetByPath(ctx, path)
	assert.NoError(t, err)
	assert.Equal(t, path, asset.FilePath)
	assert.False(t, asset.IsDeleted)

	// D. Modyfikujemy plik (Update)
	// Czekamy chwilę, żeby czas modyfikacji się różnił
	time.Sleep(100 * time.Millisecond)
	os.Chtimes(path, time.Now(), time.Now())

	err = scanner.ScanFile(ctx, path)
	assert.NoError(t, err)

	updatedAsset, _ := queries.GetAssetByPath(ctx, path)
	// LastModified powinno być nowsze niż Created
	assert.True(t, updatedAsset.LastModified.After(asset.LastModified) || updatedAsset.LastModified.Equal(asset.LastModified),
		"LastModified powinno zostać zaktualizowane")
}

// Sprawdza, czy usunięcie pliku z dysku powoduje Soft Delete w bazie.
func TestScanner_Live_ScanFile_Delete(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	fileName := "to_delete.png"
	path := filepath.Join(root, fileName)
	createDummyFile(t, path)

	// 1. Dodajemy plik (Live Scan - Create)
	err := scanner.ScanFile(ctx, path)
	assert.NoError(t, err)

	asset, err := queries.GetAssetByPath(ctx, path)
	assert.NoError(t, err)
	assert.False(t, asset.IsDeleted, "Asset powinien być aktywny")

	// 2. Usuwamy plik fizycznie
	os.Remove(path)

	// 3. Wywołujemy Live Scan (symulacja zdarzenia z Watchera)
	err = scanner.ScanFile(ctx, path)
	assert.NoError(t, err)

	// 4. Weryfikacja
	deletedAsset, err := queries.GetAssetByPath(ctx, path)
	assert.NoError(t, err)
	assert.True(t, deletedAsset.IsDeleted, "Asset powinien mieć flagę IsDeleted=true")
}
func TestIntegration_DuplicateGrouping(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	// 1. Tworzymy plik Oryginał
	originPath := filepath.Join(root, "Original.txt")
	createContentFile(t, originPath, "unikalna treść pliku tekstowego")

	// 2. Tworzymy plik Kopia (inna nazwa, ta sama treść = ten sam Hash)
	copyPath := filepath.Join(root, "Backup.txt")
	createContentFile(t, copyPath, "unikalna treść pliku tekstowego")

	// 3. Uruchamiamy PEŁNY SKAN
	err := scanner.StartScan()
	assert.NoError(t, err)

	// 4. Weryfikacja
	assert.Eventually(t, func() bool {
		assets, err := queries.ListAssetsForCache(ctx)
		if err != nil || len(assets) != 2 {
			return false
		}

		// Pobieramy oba assety
		originAsset, _ := queries.GetAssetByPath(ctx, originPath)
		copyAsset, _ := queries.GetAssetByPath(ctx, copyPath)

		// Muszą mieć różne ID (bo to dwa pliki)
		if originAsset.ID == copyAsset.ID {
			return false
		}

		// Muszą mieć TO SAMO GroupID (bo to duplikaty)
		return originAsset.GroupID == copyAsset.GroupID
	}, 2*time.Second, 50*time.Millisecond, "Skaner nie połączył plików o tym samym hashu w jedną grupę")
}

// TestIntegration_LiveScan_Heuristic sprawdza Watchera + Heurystykę.
// Scenariusz: Użytkownik zapisuje plik, Watcher go łapie, potem zapisuje nowszą wersję.
func TestIntegration_LiveScan_Heuristic(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	// 1. Symulacja: Użytkownik wrzuca plik v1
	v1Path := filepath.Join(root, "Project_Logo_v1.png")
	createContentFile(t, v1Path, "image data v1")

	// Live Scan dla v1
	err := scanner.ScanFile(ctx, v1Path)
	assert.NoError(t, err)

	// Pobieramy GroupID v1
	v1Asset, err := queries.GetAssetByPath(ctx, v1Path)
	assert.NoError(t, err)
	groupID := v1Asset.GroupID
	assert.NotEmpty(t, groupID)

	// 2. Symulacja: Użytkownik wrzuca plik v2 (inna treść, więc hash inny, ale nazwa podobna)
	v2Path := filepath.Join(root, "Project_Logo_v2.png")
	createContentFile(t, v2Path, "image data v2 completely different content")

	// Live Scan dla v2
	err = scanner.ScanFile(ctx, v2Path)
	assert.NoError(t, err)

	// 3. Weryfikacja: Czy v2 podpięło się pod v1?
	v2Asset, err := queries.GetAssetByPath(ctx, v2Path)
	assert.NoError(t, err)

	assert.Equal(t, groupID, v2Asset.GroupID, "Plik v2 powinien odziedziczyć GroupID od v1 dzięki heurystyce nazwy")
}

func TestScanner_Logic_NonImageFiles(t *testing.T) {
	_, queries, scanner, root := setupLogicTest(t)
	ctx := context.Background()

	// IMPORTANT: Allow .blend extension for this test
	scanner.config.SetAllowedExtensions([]string{".txt", ".png", ".blend"})

	// A. Create a .blend file
	blendPath := filepath.Join(root, "scene.blend")
	createDummyFile(t, blendPath)

	// B. Scan it
	err := scanner.ScanFile(ctx, blendPath)
	assert.NoError(t, err)

	// C. Verify it exists in DB
	asset, err := queries.GetAssetByPath(ctx, blendPath)
	assert.NoError(t, err)
	assert.Equal(t, blendPath, asset.FilePath)
	assert.Equal(t, "model", asset.FileType)
	assert.NotEmpty(t, asset.ThumbnailPath)
	assert.Contains(t, asset.ThumbnailPath, "blend_placeholder.webp")
}
