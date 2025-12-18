package services

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// GetFolders zwraca listę wszystkich monitorowanych folderów.
func (s *SettingsService) GetFolders() ([]database.ScanFolder, error) {
	return s.db.ListScanFolders(s.ctx)
}

// UpdateFolderStatus
func (s *SettingsService) UpdateFolderStatus(id int64, isActive bool) (database.ScanFolder, error) {
	err := s.db.UpdateScanFolderStatus(s.ctx, database.UpdateScanFolderStatusParams{
		IsActive: isActive,
		ID:       id,
	})
	if err != nil {
		return database.ScanFolder{}, err
	}
	// Jeśli wyłączamy folder, a ma on rodzica, dajmy znać że assety są "ukryte"
	if !isActive {
		folder, _ := s.db.GetScanFolderById(s.ctx, id)
		if parent := s.findBestParent(folder); parent != nil {
			wailsRuntime.EventsEmit(s.ctx, "toast", map[string]string{
				"type":    "info",
				"title":   "Monitoring Paused",
				"message": fmt.Sprintf("Assets hidden inside '%s'.", filepath.Base(parent.Path)),
			})
		}
	}

	return s.db.GetScanFolderById(s.ctx, id)
}

// DeleteFolder - KOSZ
func (s *SettingsService) DeleteFolder(id int64) error {
	targetFolder, err := s.db.GetScanFolderById(s.ctx, id)
	if err != nil {
		return err
	}
	bestParent := s.findBestParent(targetFolder)

	// Logika Reparentingu (
	if bestParent != nil {
		s.logger.Info("Deleting folder. Moving assets to parent.",
			"deleted", targetFolder.Path,
			"new_parent", bestParent.Path)

		err = s.db.MoveAssetsToFolder(s.ctx, database.MoveAssetsToFolderParams{
			ScanFolderID:   sql.NullInt64{Int64: bestParent.ID, Valid: true},
			ScanFolderID_2: sql.NullInt64{Int64: targetFolder.ID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to move assets: %w", err)
		}

		wailsRuntime.EventsEmit(s.ctx, "toast", map[string]string{
			"type":    "info",
			"title":   "Assets Saved",
			"message": fmt.Sprintf("Items moved to parent library: %s", filepath.Base(bestParent.Path)),
		})
	}
	return s.db.SoftDeleteScanFolder(s.ctx, id)
}

// AddFolder - Z obsługą przywracania (Restore)
func (s *SettingsService) AddFolder(path string) (database.ScanFolder, error) {
	if !s.ValidatePath(path) {
		return database.ScanFolder{}, errors.New("folder does not exist")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return database.ScanFolder{}, err
	}

	existing, err := s.db.GetScanFolderByPath(s.ctx, absPath)
	if err == nil {
		// ZNALEZIONO
		if existing.IsDeleted {
			s.logger.Info("Restoring folder from trash", "path", absPath)
			err := s.db.RestoreScanFolder(s.ctx, existing.ID)
			if err != nil {
				return database.ScanFolder{}, err
			}
			if !existing.IsActive {
				s.db.UpdateScanFolderStatus(s.ctx, database.UpdateScanFolderStatusParams{
					IsActive: true,
					ID:       existing.ID,
				})
			}
			return s.db.GetScanFolderById(s.ctx, existing.ID)
		}
		return database.ScanFolder{}, errors.New("folder is already in library")
	}
	return s.db.CreateScanFolder(s.ctx, absPath)
}

// Helper
func (s *SettingsService) findBestParent(target database.ScanFolder) *database.ScanFolder {
	allFolders, err := s.db.ListScanFolders(s.ctx)
	if err != nil {
		return nil
	}

	targetPath := filepath.Clean(target.Path)
	var bestParent *database.ScanFolder

	for i := range allFolders {
		f := allFolders[i]
		if f.ID == target.ID {
			continue
		}
		if !f.IsActive {
			continue
		} // Ignorujemy wyłączone

		parentPath := filepath.Clean(f.Path)
		rel, err := filepath.Rel(parentPath, targetPath)

		if err == nil && !strings.HasPrefix(rel, "..") {
			if bestParent == nil || len(f.Path) > len(bestParent.Path) {
				bestParent = &f
			}
		}
	}
	return bestParent
}

// ValidatePath sprawdza tylko czy ścieżka istnieje i jest katalogiem .
func (s *SettingsService) ValidatePath(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// OpenInExplorer otwiera menedżer plików i próbuje zaznaczyć wskazany plik/folder.
func (s *SettingsService) OpenInExplorer(path string) error {
	cleanPath := filepath.Clean(path)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", cleanPath)
	case "darwin":
		cmd = exec.Command("open", "-R", cleanPath)
	default:
		dir := filepath.Dir(cleanPath)
		if s.ValidatePath(cleanPath) {
			dir = cleanPath
		}
		cmd = exec.Command("xdg-open", dir)
	}

	return cmd.Start()
}

func (s *SettingsService) OpenFile(path string) error {
	cleanPath := filepath.Clean(path)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}

	return cmd.Start()
}
