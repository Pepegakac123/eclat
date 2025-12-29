package settings

import (
	"context"
	"database/sql"
	"eclat/internal/config"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"encoding/json"
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

// ScanFolderDTO is a Data Transfer Object for sending scan folder details to the frontend.
type ScanFolderDTO struct {
	ID          int64   `json:"id"`
	Path        string  `json:"path"`
	IsActive    bool    `json:"isActive"`
	LastScanned *string `json:"lastScanned"`
	DateAdded   string  `json:"dateAdded"`
	IsDeleted   bool    `json:"isDeleted"`
}

// AppConfigDTO is a Data Transfer Object for sending application configuration to the frontend.
type AppConfigDTO struct {
	AllowedExtensions    []string `json:"allowedExtensions"`
	MaxAllowHashFileSize int64    `json:"maxAllowHashFileSize"`
}

// WailsRuntime is an interface wrapper around Wails runtime methods to facilitate testing.
type WailsRuntime interface {
	OpenDirectoryDialog(ctx context.Context, options wailsRuntime.OpenDialogOptions) (string, error)
}

// RealWailsRuntime is the production implementation of WailsRuntime.
type RealWailsRuntime struct{}

func (r *RealWailsRuntime) OpenDirectoryDialog(ctx context.Context, options wailsRuntime.OpenDialogOptions) (string, error) {
	return wailsRuntime.OpenDirectoryDialog(ctx, options)
}

// FolderWatcher defines the interface for watching and unwatching file system directories.
type FolderWatcher interface {
	Watch(path string)
	Unwatch(path string)
}

// KeyAllowedExtensions is the database key for storing allowed file extensions.
const KeyAllowedExtensions = "allowed_extensions"

// SettingsService manages application configuration, scan folders, and system integration.
type SettingsService struct {
	ctx      context.Context
	db       database.Querier
	logger   *slog.Logger
	config   *config.ScannerConfig
	notifier feedback.Notifier
	wails    WailsRuntime
	watcher  FolderWatcher
}

// NewSettingsService creates a new instance of SettingsService.
func NewSettingsService(db database.Querier, logger *slog.Logger, notifier feedback.Notifier, watcher FolderWatcher, cfg *config.ScannerConfig) *SettingsService {
	return &SettingsService{
		db:       db,
		logger:   logger,
		notifier: notifier,
		config:   cfg,
		wails:    &RealWailsRuntime{},
		watcher:  watcher,
	}
}

// Startup is called by Wails when the application starts.
func (s *SettingsService) Startup(ctx context.Context) {
	s.ctx = ctx
	s.logger.Info("SettingsService started")
}

// --- CONFIG MANAGEMENT ---

// GetConfig returns a safe copy of the current application configuration for the UI.
func (s *SettingsService) GetConfig() AppConfigDTO {
	return AppConfigDTO{
		AllowedExtensions:    s.config.GetAllowedExtensions(),
		MaxAllowHashFileSize: s.config.GetMaxHashFileSize(),
	}
}

// SetAllowedExtensions updates the list of allowed file extensions in memory and persists it to the database.
// It validates extensions and warns about dangerous types.
func (s *SettingsService) SetAllowedExtensions(exts []string) error {
	var validExts []string
	var invalidExts []string

	for _, ext := range exts {
		if config.IsExtensionValid(ext) {
			normalized := strings.ToLower(ext)
			if !strings.HasPrefix(normalized, ".") {
				normalized = "." + normalized
			}
			validExts = append(validExts, normalized)
		} else {
			invalidExts = append(invalidExts, ext)
		}
	}
	if len(invalidExts) > 0 {
		s.logger.Warn("Attempted to add invalid extensions", "extensions", invalidExts)
		s.notifier.SendToast(s.ctx, feedback.ToastField{
			Type:    "warning",
			Title:   "Invalid Extensions Skipped",
			Message: fmt.Sprintf("Skipped dangerous or invalid types: %s", strings.Join(invalidExts, ", ")),
		})
	}
	s.config.SetAllowedExtensions(validExts)
	jsonBytes, err := json.Marshal(validExts)
	if err != nil {
		s.logger.Error("Failed to marshal extensions", "error", err)
		return fmt.Errorf("failed to save settings: %w", err)
	}

	err = s.db.SetSystemSetting(s.ctx, database.SetSystemSettingParams{
		Key:   KeyAllowedExtensions,
		Value: string(jsonBytes),
	})

	if err != nil {
		s.logger.Error("Failed to persist settings to DB", "error", err)
	} else {
		s.logger.Info("Extensions saved to DB", "count", len(validExts))
	}

	return nil
}

// --- FOLDER MANAGEMENT ---

// GetFolders retrieves the list of configured scan folders.
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

// UpdateFolderStatus toggles the active state of a folder.
// If a folder is deactivated, its assets are hidden from the library.
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

	statusMsg := "paused"
	if isActive {
		statusMsg = "active"
	}

	s.notifier.SendToast(s.ctx, feedback.ToastField{
		Type:    "info",
		Title:   "Folder Status Updated",
		Message: fmt.Sprintf("Folder is now %s. Please run a scan to update your library.", statusMsg),
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

// DeleteFolder soft-deletes a scan folder.
// Before deletion, it attempts to move assets to a parent folder if one exists in the library,
// preserving the assets' history.
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
		s.notifier.SendToast(s.ctx, feedback.ToastField{
			Type:    "info",
			Title:   "Assets Saved",
			Message: fmt.Sprintf("Items moved to parent library: %s", filepath.Base(bestParent.Path)),
		})

	}
	if s.watcher != nil && targetFolder.Path != "" {
		s.watcher.Unwatch(targetFolder.Path)
	}
	err = s.db.SoftDeleteScanFolder(s.ctx, id)
	if err == nil {
		s.notifier.SendToast(s.ctx, feedback.ToastField{
			Type:    "info",
			Title:   "Folder Removed",
			Message: "Folder removed. Please run a scan to cleanup the library.",
		})
	}
	return err
}

// AddFolder adds a new directory to the list of scan folders.
// If the folder was previously deleted, it restores it.
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
			if s.watcher != nil {
				s.watcher.Watch(absPath)
			}
			restored, _ := s.db.GetScanFolderById(s.ctx, existing.ID)

			s.notifier.SendToast(s.ctx, feedback.ToastField{
				Type:    "success",
				Title:   "Folder Re-added",
				Message: "This folder was previously removed. Its monitoring has been resumed. Please run a scan to update your library.",
			})

			return s.mapToDTO(restored), nil
		}
		return ScanFolderDTO{}, errors.New("folder is already in library")
	}

	newFolder, err := s.db.CreateScanFolder(s.ctx, absPath)
	if err != nil {
		return ScanFolderDTO{}, err
	}
	if s.watcher != nil {
		s.watcher.Watch(absPath)
	}

	s.notifier.SendToast(s.ctx, feedback.ToastField{
		Type:    "success",
		Title:   "Folder Added",
		Message: "New source folder added. Please run a scan to import assets.",
	})

	return s.mapToDTO(newFolder), nil
}

// mapToDTO converts a database ScanFolder model to a frontend DTO.
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
		LastScanned: lastScannedStr,
		DateAdded:   f.DateAdded.Format(time.RFC3339),
		IsDeleted:   f.IsDeleted,
	}
}

// findBestParent locates a parent folder in the current library for a given folder.
// This is used when deleting a subfolder to check if its contents are covered by a root folder.
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

// ValidatePath checks if a given path exists and is a directory.
func (s *SettingsService) ValidatePath(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// OpenInExplorer opens the system file explorer at the specified path.
// It supports Windows, macOS, and Linux.
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

// OpenFile opens the specified file in its default associated application.
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

// OpenFolderPicker opens a native system dialog to select a directory.
func (s *SettingsService) OpenFolderPicker() (string, error) {
	selection, err := s.wails.OpenDirectoryDialog(s.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Library Folder",
	})
	if err != nil {
		return "", err
	}

	return selection, nil
}
