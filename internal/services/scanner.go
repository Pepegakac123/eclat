package services

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

/*
TODO: GAP ANALYSIS - SCANNER IMPLEMENTATION

1. LOGIKA BIZNESOWA (SYNCHRONIZACJA):
   - [ ] Check Existing: Przed insertem sprawdź, czy plik o tej ścieżce istnieje w DB.
   - [ ] Self-Healing: Jeśli pliku nie ma pod ścieżką, ale zgadza się Hash (dla przeniesionych plików) -> Zaktualizuj ścieżkę i ScanFolderID zamiast tworzyć duplikat.
   - [ ] Soft Delete: Wykrywanie plików usuniętych z dysku lub takich, których rozszerzenie przestało być dozwolone (ustaw flagę IsDeleted).
   - [ ] Restore: Przywracanie plików (usuwanie flagi IsDeleted), jeśli plik wrócił lub rozszerzenie znów jest dozwolone.
   - [ ] Ignore Large Files: Pomijanie hashowania dla plików > X MB (dla wydajności).

2. PRZETWARZANIE MEDIÓW (C# używa ImageSharp):
   - [ ] Thumbnails: Integracja biblioteki do skalowania obrazów (np. "github.com/disintegration/imaging").
   - [ ] Metadata: Wyciąganie wymiarów (width/height) i głębi kolorów.
   - [ ] Video/3D: Obsługa placeholderów dla plików, których nie umiemy otworzyć (np. .blend, .fbx).

3. WYDAJNOŚĆ I CONCURRENCY:
   - [ ] Worker Pool: Zastąpienie sekwencyjnego `WalkDir` wzorcem Producer-Consumer.
         (WalkDir wrzuca ścieżki na kanał -> N gorutyn przetwarza pliki równolegle).
   - [ ] Batch Insert: Zbieranie wyników i zapisywanie do DB w transakcjach po 100-500 sztuk (znacznie szybsze niż pojedyncze INSERT).
   - [ ] Streaming Hash: Obliczanie SHA256 przy użyciu `io.Copy` (małe zużycie RAM), a nie wczytywanie całego pliku.

4. UI FEEDBACK:
   - [ ] Progress Throttling: Wysyłanie eventów do UI nie częściej niż co X milisekund (aby nie zamrozić Frontendu).
*/

// Scanner is a struct that keeps needed dependencies for scanning assets.
type Scanner struct {
	mu         sync.RWMutex
	db         *database.Queries
	logger     *slog.Logger
	cancelFunc context.CancelFunc
	config     *ScannerConfig
	isScanning atomic.Bool
}

// NewScanner creates a new Scanner instance
func NewScanner(db *database.Queries) *Scanner {
	return &Scanner{
		db:     db,
		logger: slog.Default(),
		config: NewScannerConfig(),
	}
}

type CachedAsset struct {
	ID           int64
	LastModified time.Time
}

// StartScan starts the scanning process in the background
func (s *Scanner) StartScan(wailsCtx context.Context) error {
	if s.isScanning.Load() {
		return nil
	}
	scanCtx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	s.isScanning.Store(true)

	go func() {
		defer s.isScanning.Store(false)
		defer cancel()

		s.logger.Info("Scanner Started")

		// Get active folders
		folders, err := s.db.ListScanFolders(scanCtx)
		if err != nil {
			s.logger.Error("Failed to list folders", slog.String("error", err.Error()))
		}

		s.logger.Info("Calculating total files...")
		totalToProcess := 0
		for _, f := range folders {
			if scanCtx.Err() != nil {
				s.logger.Info("Scanner cancelled by user")
				break
			}
			totalToProcess += s.getAllFilesCount(f)
		}
		s.logger.Info("Total files to scan calculated", "count", totalToProcess)

		existingAssets, err := s.loadExistingAssets(scanCtx)
		if err != nil {
			s.logger.Error("Failed to load asset cache", "error", err)
			return
		}

		totalProcessed := 0
		runtime.EventsEmit(wailsCtx, "scan_status", "scanning") // Needed for UI Scanner Status Update
		for _, f := range folders {
			// Check for the cancellation
			if scanCtx.Err() != nil {
				s.logger.Info("Scanner cancelled by user")
				break
			}
			s.scanDirectory(scanCtx, wailsCtx, f, &totalProcessed, totalToProcess, existingAssets)
		}
		s.logger.Info("Scanner finished", "total", totalProcessed)
		runtime.EventsEmit(wailsCtx, "scan_status", "idle")
	}()
	return nil

}

// StopScan cancells active project
func (s *Scanner) StopScan() {
	if s.cancelFunc != nil {
		s.cancelFunc() // Sends done() signal to goroutine
	}
}

// Helper function that scans a directory for files, Add them to db and updates the total count
func (s *Scanner) scanDirectory(ctx context.Context, wailsCtx context.Context, folder database.ScanFolder, total *int, totalToProcess int, existingCache map[string]CachedAsset) {
	err := filepath.WalkDir(folder.Path, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return filepath.SkipAll
		}
		if err != nil {
			s.logger.Warn("Error accessing path", "path", path, "error", err.Error())
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if isAllowed := s.IsExtensionAllowed(ext); !isAllowed {
			s.logger.Debug("Extension not allowed", "path", path)
			return nil
		}
		info, _ := d.Info()
		if _, exists := existingCache[path]; exists {
			cachedTime := existingCache[path].LastModified
			diskTime := info.ModTime()
			if !diskTime.Equal(cachedTime) {
				//TODO: Update logic, robimy checka nazwy pliku i innych danych
				// s.logger.Info("File modified, updating...", "path", path)
			}
			// Update procesu skanowania
			s.updateAndEmitTotal(total, d, wailsCtx, totalToProcess)
			//Tak to skip
			return nil
		}

		// TODO: Add thumbnail generation logic

		_, dbErr := s.db.CreateAsset(ctx, database.CreateAssetParams{
			ScanFolderID:  sql.NullInt64{Int64: int64(folder.ID), Valid: true},
			FileName:      d.Name(),
			FilePath:      path,
			FileType:      s.determineFileType(ext),
			FileSize:      info.Size(),
			ThumbnailPath: "", // TODO: Generate thumbnail logic and path
			LastModified:  info.ModTime(),
			LastScanned:   time.Now(),
			DominantColor: s.getDominantColor(path),
			// TODO: Add more fields as needed
		})
		if dbErr != nil {
			// TODO: Handle UNIQUE constraint violation
			s.logger.Debug("Skipping asset", "path", path, "reason", dbErr)
		} else {
			s.updateAndEmitTotal(total, d, wailsCtx, totalToProcess)
		}

		return nil
	})
	if err != nil {
		s.logger.Error("WalkDir failed", "path", folder.Path, "error", err)
	}
}

func (s *Scanner) getAllFilesCount(folder database.ScanFolder) int {
	var total = 0
	err := filepath.WalkDir(folder.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if s.IsExtensionAllowed(ext) {
			total++
		}
		return nil
	})
	if err != nil {
		s.logger.Error("WalkDir failed", "path", folder.Path, "error", err)
	}
	return total
}

func (s *Scanner) loadExistingAssets(ctx context.Context) (map[string]CachedAsset, error) {
	s.logger.Info("Loading assets paths to the memmory")

	rows, err := s.db.ListAssetsPath(ctx)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]CachedAsset, len(rows))

	for _, row := range rows {
		existing[row.FilePath] = CachedAsset{
			ID:           row.ID,
			LastModified: row.LastModified,
		}
	}
	s.logger.Info("Loaded assets cache", "count", len(existing))
	return existing, nil

}

func (s *Scanner) updateAndEmitTotal(total *int, d fs.DirEntry, wailsCtx context.Context, totalToProcess int) {
	*total++
	if *total%10 == 0 {
		runtime.EventsEmit(wailsCtx, "scan_progress", map[string]any{
			"current":  *total,
			"total":    totalToProcess,
			"lastFile": d.Name(),
		})
	}
}
