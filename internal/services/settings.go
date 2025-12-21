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
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ScanFolderDTO - Struktura bezpieczna dla Wails/Frontend
type ScanFolderDTO struct {
	ID          int64   `json:"id"`
	Path        string  `json:"path"`
	IsActive    bool    `json:"isActive"`
	LastScanned *string `json:"lastScanned"`
	DateAdded   string  `json:"dateAdded"`
	IsDeleted   bool    `json:"isDeleted"`
}
type WailsRuntime interface {
	OpenDirectoryDialog(ctx context.Context, options wailsRuntime.OpenDialogOptions) (string, error)
}

type RealWailsRuntime struct{}

func (r *RealWailsRuntime) OpenDirectoryDialog(ctx context.Context, options wailsRuntime.OpenDialogOptions) (string, error) {
	return wailsRuntime.OpenDirectoryDialog(ctx, options)
}

// SettingsService odpowiada za konfigurację aplikacji i zarządzanie biblioteką.
type SettingsService struct {
	ctx      context.Context
	db       database.Querier
	logger   *slog.Logger
	notifier Notifier
	wails    WailsRuntime
}

// NewSettingsService tworzy nową instancję serwisu.
func NewSettingsService(db database.Querier, logger *slog.Logger, notifier Notifier) *SettingsService {
	return &SettingsService{
		db:       db,
		logger:   logger,
		notifier: notifier,
		wails:    &RealWailsRuntime{},
	}
}

// Startup jest wywoływany przez Wails przy starcie aplikacji.
func (s *SettingsService) Startup(ctx context.Context) {
	s.ctx = ctx
	s.logger.Info("SettingsService started")
}

func (s *SettingsService) GetFolders() ([]ScanFolderDTO, error) {
	folders, err := s.db.ListScanFolders(s.ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]ScanFolderDTO, len(folders))
	for i, f := range folders {
		dtos[i] = s.mapToDTO(f)
	}

	return dtos, nil
}

// UpdateFolderStatus toggles the active state of a folder and updates visibility of its assets.
func (s *SettingsService) UpdateFolderStatus(id int64, isActive bool) (ScanFolderDTO, error) {

	err := s.db.UpdateScanFolderStatus(s.ctx, database.UpdateScanFolderStatusParams{
		IsActive: isActive,
		ID:       id,
	})
	if err != nil {
		s.logger.Error("Failed to update folder status", "error", err)
		return ScanFolderDTO{}, err
	}

	shouldHide := !isActive

	err = s.db.SetAssetsHiddenByFolderId(s.ctx, database.SetAssetsHiddenByFolderIdParams{
		IsHidden:     shouldHide,
		ScanFolderID: sql.NullInt64{Int64: id, Valid: true},
	})

	if err != nil {
		s.logger.Error("Failed to update assets visibility", "folderId", id, "error", err)
	}

	// C. UI Feedback (Toast)
	statusMsg := "restored"
	if !isActive {
		statusMsg = "hidden"
	}

	s.notifier.SendToast(s.ctx, ToastField{
		Type:    "info",
		Title:   "Folder Updated",
		Message: fmt.Sprintf("Folder is now %s. Assets are %s.", boolToStatus(isActive), statusMsg),
	})
	updatedFolder, err := s.db.GetScanFolderById(s.ctx, id)
	if err != nil {
		return ScanFolderDTO{}, err
	}
	return s.mapToDTO(updatedFolder), nil
}
func boolToStatus(active bool) string {
	if active {
		return "Active"
	}
	return "Paused"
}

// DeleteFolder - KOSZ
func (s *SettingsService) DeleteFolder(id int64) error {
	targetFolder, err := s.db.GetScanFolderById(s.ctx, id)
	if err != nil {
		return err
	}
	bestParent := s.findBestParent(targetFolder)

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
		s.notifier.SendToast(s.ctx, ToastField{
			Type:    "info",
			Title:   "Assets Saved",
			Message: fmt.Sprintf("Items moved to parent library: %s", filepath.Base(bestParent.Path)),
		})

	}
	return s.db.SoftDeleteScanFolder(s.ctx, id)
}

// AddFolder - Z obsługą przywracania (Restore)
// --- ZMIANA: Zwracamy ScanFolderDTO ---
func (s *SettingsService) AddFolder(path string) (ScanFolderDTO, error) {
	if !s.ValidatePath(path) {
		return ScanFolderDTO{}, errors.New("folder does not exist")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return ScanFolderDTO{}, err
	}

	existing, err := s.db.GetScanFolderByPath(s.ctx, absPath)
	if err == nil {
		// ZNALEZIONO
		if existing.IsDeleted {
			s.logger.Info("Restoring folder from trash", "path", absPath)
			err := s.db.RestoreScanFolder(s.ctx, existing.ID)
			if err != nil {
				return ScanFolderDTO{}, err
			}
			if !existing.IsActive {
				s.db.UpdateScanFolderStatus(s.ctx, database.UpdateScanFolderStatusParams{
					IsActive: true,
					ID:       existing.ID,
				})
			}
			restored, _ := s.db.GetScanFolderById(s.ctx, existing.ID)
			return s.mapToDTO(restored), nil
		}
		return ScanFolderDTO{}, errors.New("folder is already in library")
	}

	newFolder, err := s.db.CreateScanFolder(s.ctx, absPath)
	if err != nil {
		return ScanFolderDTO{}, err
	}
	return s.mapToDTO(newFolder), nil
}

// Helper: mapToDTO konwertuje struct bazy na struct dla Frontendu
func (s *SettingsService) mapToDTO(f database.ScanFolder) ScanFolderDTO {
	var lastScannedStr *string
	if f.LastScanned.Valid {
		formatted := f.LastScanned.Time.Format(time.RFC3339)
		lastScannedStr = &formatted
	}

	return ScanFolderDTO{
		ID:          f.ID,
		Path:        f.Path,
		IsActive:    f.IsActive,
		LastScanned: lastScannedStr,                   // Teraz to *string
		DateAdded:   f.DateAdded.Format(time.RFC3339), // Teraz to string
		IsDeleted:   f.IsDeleted,
	}
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
		}

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

func (s *SettingsService) ValidatePath(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

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

// OpenFolderPicker otwiera systemowe okno, które widzi TYLKO foldery
func (s *SettingsService) OpenFolderPicker() (string, error) {
	selection, err := s.wails.OpenDirectoryDialog(s.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Library Folder",
	})
	if err != nil {
		return "", err
	}

	return selection, nil
}
