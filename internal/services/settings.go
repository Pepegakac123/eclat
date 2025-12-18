package services

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SettingsService odpowiada za konfigurację aplikacji i zarządzanie biblioteką.
type SettingsService struct {
	ctx    context.Context
	db     *database.Queries
	logger *slog.Logger
}

// NewSettingsService tworzy nową instancję serwisu.
func NewSettingsService(db *database.Queries) *SettingsService {
	return &SettingsService{
		db:     db,
		logger: slog.Default(),
	}
}

// Startup jest wywoływany przez Wails przy starcie aplikacji.
func (s *SettingsService) Startup(ctx context.Context) {
	s.ctx = ctx
	s.logger.Info("SettingsService started")
}

// --- SEKJA 1: Zarządzanie Folderami (Scan Folders) ---

// GetFolders zwraca listę wszystkich monitorowanych folderów.
func (s *SettingsService) GetFolders() ([]database.ScanFolder, error) {
	// TODO: Użyj s.db.ListScanFolders(s.ctx)
	// Pamiętaj o obsłudze błędów.
	return nil, nil
}

// AddFolder dodaje nowy folder do bazy.
// Wymagania (z SettingsRepository.cs):
// 1. Sprawdź czy folder istnieje fizycznie na dysku (os.Stat).
// 2. Sprawdź czy folder już jest w bazie (GetScanFolderByPath).
// 3. Jeśli jest w bazie i ma flagę is_deleted=1 -> przywróć go (RestoreScanFolder).
// 4. Jeśli nie ma -> utwórz nowy (CreateScanFolder).
func (s *SettingsService) AddFolder(path string) (database.ScanFolder, error) {
	// TODO: Twoja implementacja tutaj.
	// Wskazówka: filepath.Abs(path) jest Twoim przyjacielem.
	return database.ScanFolder{}, nil
}

func (s *SettingsService) DeleteFolder(id int64) error {
	ctx := context.Background()

	targetFolder, err := s.db.GetScanFolderById(ctx, id)
	if err != nil {
		return err
	}

	allFolders, err := s.db.ListScanFolders(ctx)
	if err != nil {
		return err
	}
	var bestParent *database.ScanFolder
	targetPath := filepath.Clean(targetFolder.Path)

	for _, f := range allFolders {
		if f.ID == targetFolder.ID {
			continue
		} // Pomiń samego siebie

		parentPath := filepath.Clean(f.Path)

		// Sprawdź czy parentPath jest prefixem targetPath
		// Np. Parent: D:\Docs, Target: D:\Docs\Assets -> TAK
		rel, err := filepath.Rel(parentPath, targetPath)
		if err == nil && !strings.HasPrefix(rel, "..") {
			// Jest w środku! Sprawdź czy to "najgłębszy" rodzic
			if bestParent == nil || len(f.Path) > len(bestParent.Path) {
				temp := f
				bestParent = &temp
			}
		}
	}

	// 4. Decyzja
	if bestParent != nil {
		s.logger.Info("Deleting folder but parent exists. Re-binding assets.",
			"deleted", targetFolder.Path,
			"new_parent", bestParent.Path)

		// A. Przepnij assety do rodzica
		err = s.db.MoveAssetsToFolder(ctx, database.MoveAssetsToFolderParams{
			ScanFolderID:   sql.NullInt64{Int64: bestParent.ID, Valid: true},
			ScanFolderID_2: sql.NullInt64{Int64: targetFolder.ID, Valid: true},
		})
		if err != nil {
			return err
		}
		runtime.EventsEmit(s.ctx, "toast", map[string]string{
			"type":    "info",
			"title":   "Assets Re-organized",
			"message": fmt.Sprintf("Folder removed, but assets were moved to parent: %s", filepath.Base(bestParent.Path)),
		})

	} else {
		s.logger.Info("Deleting folder. No parent found. Assets will be hidden.")
	}
	return s.db.SoftDeleteScanFolder(ctx, id)
}

// UpdateFolderStatus zmienia status aktywności folderu (włącz/wyłącz skanowanie).
func (s *SettingsService) UpdateFolderStatus(id int64, isActive bool) (database.ScanFolder, error) {
	// TODO:
	// 1. Wykonaj s.db.UpdateScanFolderStatus
	// 2. Pobierz zaktualizowany obiekt (np. przez GetScanFolderByID - musisz dodać to query jeśli nie masz,
	//    lub po prostu zwróć skonstruowany obiekt jeśli lenistwo wygra).
	return database.ScanFolder{}, nil
}

// ValidatePath sprawdza tylko czy ścieżka istnieje i jest katalogiem (dla formularza UI).
func (s *SettingsService) ValidatePath(path string) bool {
	// TODO: Użyj os.Stat. Zwróć true tylko jeśli err == nil i FileInfo.IsDir() jest true.
	return false
}

// OpenInExplorer otwiera systemowy eksplorator plików na danej ścieżce.
func (s *SettingsService) OpenInExplorer(path string) error {
	// TODO: Wykorzystaj runtime.GOOS i exec.Command ("explorer", "open", "xdg-open").
	return nil
}
