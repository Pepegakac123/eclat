package settings

import (
	"context"
	"database/sql"
	"eclat/internal/config"
	"eclat/internal/database"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// MockWailsRuntime do testowania okien dialogowych bez UI
type MockWailsRuntime struct {
	SelectedPath string
	ShouldError  bool
}

func (m *MockWailsRuntime) OpenDirectoryDialog(ctx context.Context, options wailsRuntime.OpenDialogOptions) (string, error) {
	if m.ShouldError {
		return "", os.ErrPermission
	}
	return m.SelectedPath, nil
}

type NoOpWatcher struct{}

func (nw *NoOpWatcher) Watch(path string)   {}
func (nw *NoOpWatcher) Unwatch(path string) {}

func TestSettings_ValidatePath(t *testing.T) {
	mockNotifier := &MockNotifier{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	logLevel := &slog.LevelVar{}
	// Tutaj nil dla DB jest OK, bo ValidatePath używa tylko os.Stat
	svc := NewSettingsService(nil, logger, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())

	t.Run("Should return true for existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		isValid := svc.ValidatePath(tempDir)
		assert.True(t, isValid)
	})

	t.Run("Should return false for non-existing path", func(t *testing.T) {
		isValid := svc.ValidatePath("/sciezka/ktora/nie/istnieje/999")
		assert.False(t, isValid)
	})

	t.Run("Should return false for a file (not directory)", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		isValid := svc.ValidatePath(tempFile.Name())
		assert.False(t, isValid, "Plik nie powinien być rozpoznany jako folder")
	})
}

func TestSettings_ConfigUpdates(t *testing.T) {
	// POPRAWKA 1: Musimy mieć bazę danych, bo SetAllowedExtensions zapisuje do niej!
	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockNotifier := &MockNotifier{}

	cfg := config.NewScannerConfig()
	logLevel := &slog.LevelVar{}
	// POPRAWKA 2: Przekazujemy queries zamiast nil
	svc := NewSettingsService(queries, logger, logLevel, mockNotifier, &NoOpWatcher{}, cfg)

	// POPRAWKA 3: Inicjalizujemy kontekst, bo baza go wymaga
	ctx := context.Background()
	svc.Startup(ctx)

	t.Run("Should update extensions in shared config", func(t *testing.T) {
		// Początkowy stan
		initial := svc.GetConfig()
		assert.Contains(t, initial.AllowedExtensions, ".jpg")

		// Zmiana
		newExts := []string{".blend", ".fbx", "OBJ"} // OBJ bez kropki - powinno naprawić
		err := svc.SetAllowedExtensions(newExts)
		assert.NoError(t, err)

		// Sprawdzamy przez serwis
		updated := svc.GetConfig()
		assert.Len(t, updated.AllowedExtensions, 3)
		assert.Contains(t, updated.AllowedExtensions, ".obj")

		// KLUCZOWE: Sprawdzamy czy "shared config" został zaktualizowany
		assert.True(t, cfg.IsExtensionAllowed(".blend"))
		assert.False(t, cfg.IsExtensionAllowed(".jpg"), "Stare rozszerzenie powinno zniknąć")

		// Sprawdźmy czy zapisało się w bazie (Persistence Check)
		storedJSON, err := queries.GetSystemSetting(ctx, "allowed_extensions")
		assert.NoError(t, err)
		assert.Contains(t, storedJSON, ".blend", "Ustawienia powinny trafić do bazy")
	})

	t.Run("Should filter dangerous extensions", func(t *testing.T) {
		badExts := []string{".exe", ".bat", ".png"}
		err := svc.SetAllowedExtensions(badExts)
		assert.NoError(t, err)

		cfgDto := svc.GetConfig()
		assert.Contains(t, cfgDto.AllowedExtensions, ".png")
		assert.NotContains(t, cfgDto.AllowedExtensions, ".exe")
		assert.NotContains(t, cfgDto.AllowedExtensions, ".bat")
	})
}

func TestSettings_ScanFolder_CRUD_FullFlow(t *testing.T) {
	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockNotifier := &MockNotifier{}
	logLevel := &slog.LevelVar{}
	svc := NewSettingsService(queries, logger, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())
	ctx := context.Background()
	svc.Startup(ctx)

	folderPath := t.TempDir()

	t.Run("Add and Get Folder", func(t *testing.T) {
		added, err := svc.AddFolder(folderPath)
		assert.NoError(t, err)
		assert.Equal(t, folderPath, added.Path)

		folders, err := svc.GetFolders()
		assert.NoError(t, err)
		assert.Len(t, folders, 1)
	})

	t.Run("Reject Duplicate Path", func(t *testing.T) {
		_, err := svc.AddFolder(folderPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already in library")
	})

	t.Run("Update Status", func(t *testing.T) {
		folders, _ := svc.GetFolders()
		id := folders[0].ID

		updated, err := svc.UpdateFolderStatus(id, false)
		assert.NoError(t, err)
		assert.False(t, updated.IsActive)
	})

	t.Run("Soft Delete Folder", func(t *testing.T) {
		folders, _ := svc.GetFolders()
		assert.NotEmpty(t, folders)
		id := folders[0].ID

		err := svc.DeleteFolder(id)
		assert.NoError(t, err)

		foldersAfter, _ := svc.GetFolders()
		foundInDTO := false
		for _, f := range foldersAfter {
			if f.ID == id {
				foundInDTO = true
				break
			}
		}
		assert.False(t, foundInDTO, "Folder powinien zostać odfiltrowany z listy (is_deleted=1)")
		dbFolder, err := queries.GetScanFolderById(ctx, id)
		assert.NoError(t, err)
		assert.True(t, dbFolder.IsDeleted, "Rekord w bazie danych powinien mieć IsDeleted = 1")
	})
}

func TestSettings_FolderPicker_Mock(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockNotifier := &MockNotifier{}
	logLevel := &slog.LevelVar{}
	svc := NewSettingsService(nil, logger, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())
	svc.Startup(context.Background())

	t.Run("Successful Selection", func(t *testing.T) {
		mockPath := "/fake/path/to/library"
		svc.wails = &MockWailsRuntime{SelectedPath: mockPath}

		path, err := svc.OpenFolderPicker()
		assert.NoError(t, err)
		assert.Equal(t, mockPath, path)
	})
}

func TestSettings_FindBestParent_Logic(t *testing.T) {
	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockNotifier := &MockNotifier{}
	logLevel := &slog.LevelVar{}
	svc := NewSettingsService(queries, logger, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())
	ctx := context.Background()
	svc.Startup(ctx)

	rootPath := "/tmp/eclat_root"
	subPath := filepath.Join(rootPath, "sub")

	queries.CreateScanFolder(ctx, rootPath)
	subFolder, _ := queries.CreateScanFolder(ctx, subPath)

	t.Run("Find parent of subfolder", func(t *testing.T) {
		parent := svc.findBestParent(subFolder)
		assert.NotNil(t, parent)
		assert.Equal(t, rootPath, parent.Path)
	})
}

func TestSettings_MapToDTO_Scenarios(t *testing.T) {
	mockNotifier := &MockNotifier{}
	logLevel := &slog.LevelVar{}
	svc := NewSettingsService(nil, nil, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())

	t.Run("Map with LastScanned NULL", func(t *testing.T) {
		f := database.ScanFolder{LastScanned: sql.NullTime{Valid: false}, DateAdded: time.Now()}
		dto := svc.mapToDTO(f)
		assert.Nil(t, dto.LastScanned)
	})

	t.Run("Map with LastScanned Value", func(t *testing.T) {
		now := time.Now()
		f := database.ScanFolder{LastScanned: sql.NullTime{Time: now, Valid: true}, DateAdded: now}
		dto := svc.mapToDTO(f)
		assert.NotNil(t, dto.LastScanned)
		assert.Equal(t, now.Format(time.RFC3339), *dto.LastScanned)
	})
}

func TestSettings_UpdateFolderStatus_Logic(t *testing.T) {
	_, queries, _, root := setupLogicTest(t)
	ctx := context.Background()
	folders, _ := queries.ListScanFolders(ctx)
	folderID := folders[0].ID

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockNotifier := &MockNotifier{}
	logLevel := &slog.LevelVar{}
	svc := NewSettingsService(queries, logger, logLevel, mockNotifier, &NoOpWatcher{}, config.NewScannerConfig())
	svc.Startup(ctx)
	path := filepath.Join(root, "test.png")
	createDummyFile(t, path)
	insertTestAsset(t, queries, folderID, path, "hash123")
	_, err := svc.UpdateFolderStatus(folderID, false)
	assert.NoError(t, err)

	asset, _ := queries.GetAssetByPath(ctx, path)
	assert.True(t, asset.IsHidden, "Asset powinien zostać ukryty wraz z folderem")

	svc.UpdateFolderStatus(folderID, true)
	asset, _ = queries.GetAssetByPath(ctx, path)
	assert.False(t, asset.IsHidden, "Asset powinien zostać odkryty")
}
