package services

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

/*
TODO: GAP ANALYSIS - SCANNER IMPLEMENTATION

3. WYDAJNOŚĆ I CONCURRENCY:
   - [ ] Worker Pool: Zastąpienie sekwencyjnego `WalkDir` wzorcem Producer-Consumer.
         (WalkDir wrzuca ścieżki na kanał -> N gorutyn przetwarza pliki równolegle).
   - [ ] Batch Insert: Zbieranie wyników i zapisywanie do DB w transakcjach po 100-500 sztuk (znacznie szybsze niż pojedyncze INSERT).

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
	thumbGen   *ThumbnailGenerator
	ctx        context.Context
}
type ScanJob struct {
	Path     string
	FolderId int64
	Entry    fs.DirEntry
}
type ScanResult struct {
	Path       string
	Err        error
	NewAsset   *database.CreateAssetParams
	ExistingID int64
}

func (s *Scanner) Startup(ctx context.Context) {
	s.ctx = ctx
}

// NewScanner creates a new Scanner instance
func NewScanner(db *database.Queries, thumbGen *ThumbnailGenerator) *Scanner {
	return &Scanner{
		db:       db,
		logger:   slog.Default(),
		config:   NewScannerConfig(),
		thumbGen: thumbGen,
	}
}

type CachedAsset struct {
	ID           int64
	LastModified time.Time
	IsDeleted    bool
	ScanFolderID sql.NullInt64
}

// StartScan starts the scanning process in the background
func (s *Scanner) StartScan() error {
	if s.isScanning.Load() {
		return nil
	}
	// kontekst do anulowania samej pracy (background work)
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
		foundAssets := make(map[int64]bool)

		totalProcessed := 0
		runtime.EventsEmit(s.ctx, "scan_status", "scanning") // Needed for UI Scanner Status Update
		for _, f := range folders {
			// Check for the cancellation
			if scanCtx.Err() != nil {
				s.logger.Info("Scanner cancelled by user")
				break
			}
			s.scanDirectory(scanCtx, f, &totalProcessed, totalToProcess, existingAssets, foundAssets)
		}
		s.logger.Info("Scanner finished", "total", totalProcessed)
		runtime.EventsEmit(s.ctx, "scan_status", "idle")
		s.logger.Info("Starting Cleanup Phase (Soft Delete)...")
		for path, cached := range existingAssets {
			if !foundAssets[cached.ID] {
				if !cached.IsDeleted {
					s.logger.Info("Asset missing or invalid extension - Soft Deleting", "path", path)
					s.db.SoftDeleteAsset(scanCtx, cached.ID)
				}
			}
		}
	}()
	return nil

}

// StopScan cancells active project
func (s *Scanner) StopScan() {
	if s.cancelFunc != nil {
		s.cancelFunc() // Sends done() signal to goroutine
	}
}
func (s *Scanner) Worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan ScanJob, results chan<- ScanResult) {
	defer wg.Done()

	for job := range jobs {
		result := ScanResult{Path: job.Path}
		path := job.Path
		folderId := job.FolderId
		entry := job.Entry

		ext := filepath.Ext(path)
		if isAllowed := s.IsExtensionAllowed(ext); !isAllowed {
			s.logger.Debug("Extension not allowed", "path", path)
			continue
		}

		fileType := DetermineFileType(ext)

		hash, err := CalculateFileHash(path, s.config.MaxAllowHashFileSize)
		if err != nil {
			s.logger.Debug("Failed to calculate hash", "path", path, "error", err)
			result.Err = err
			results <- result
			continue
		}

		var exist database.Asset
		var lookupErr error

		if hash != "" {
			// Strategia A: Mamy hash, szukamy po hashu
			exist, lookupErr = s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
		} else {
			// Strategia B: Brak hasha (duży plik), szukamy po ścieżce
			exist, lookupErr = s.db.GetAssetByPath(ctx, path)
		}

		if lookupErr != nil && lookupErr != sql.ErrNoRows {
			s.logger.Debug("DB Lookup failed", "path", path, "error", lookupErr)
			result.Err = lookupErr
			results <- result
			continue
		}
		if exist.ID > 0 {
			result.ExistingID = exist.ID
		} else {
			if lookupErr == sql.ErrNoRows {
				s.logger.Debug("Asset not found (new file)", "path", path)
			}
			newAsset, err := s.generateAssetMetadata(ctx, path, entry, folderId, fileType, hash)
			if err != nil {
				s.logger.Debug("Failed to generate metadata", "path", path, "error", err)
				result.Err = err
			} else {
				result.NewAsset = &newAsset
			}
		}

		results <- result
	}
}

// Funkcja pomocnicza -
func (s *Scanner) generateAssetMetadata(ctx context.Context, path string, entry fs.DirEntry, folderId int64, filetype string, hash string) (database.CreateAssetParams, error) {
	thumb, err := s.thumbGen.Generate(ctx, path)
	if err != nil {
		s.logger.Debug("Failed to generate thumbnail", "path", path, "error", err)
		return database.CreateAssetParams{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		s.logger.Debug("Failed to get file info", "path", path, "error", err)
		return database.CreateAssetParams{}, err
	}

	hasValidDimensions := thumb.Metadata.Width > 0 && thumb.Metadata.Height > 0

	newAsset := database.CreateAssetParams{
		ScanFolderID:    sql.NullInt64{Int64: folderId, Valid: true},
		FileName:        entry.Name(),
		FilePath:        path,
		FileType:        filetype,
		FileSize:        info.Size(),
		ThumbnailPath:   thumb.WebPath,
		FileHash:        sql.NullString{String: hash, Valid: hash != ""},
		ImageWidth:      sql.NullInt64{Int64: int64(thumb.Metadata.Width), Valid: hasValidDimensions},
		ImageHeight:     sql.NullInt64{Int64: int64(thumb.Metadata.Height), Valid: hasValidDimensions},
		DominantColor:   sql.NullString{String: string(thumb.Metadata.DominantColor), Valid: thumb.Metadata.DominantColor != ""},
		BitDepth:        sql.NullInt64{Int64: int64(thumb.Metadata.BitDepth), Valid: hasValidDimensions},
		HasAlphaChannel: sql.NullBool{Bool: thumb.Metadata.HasAlphaChannel, Valid: hasValidDimensions},
		LastModified:    info.ModTime(),
		LastScanned:     time.Now(),
	}

	return newAsset, nil
}

// Helper function that scans a directory for files, Add them to db and updates the total count
func (s *Scanner) scanDirectory(ctx context.Context, folder database.ScanFolder, total *int, totalToProcess int, existingCache map[string]CachedAsset, foundAssets map[int64]bool) {
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
		fileType := DetermineFileType(ext)
		if cached, exists := existingCache[path]; exists {
			foundAssets[cached.ID] = true

			if cached.ScanFolderID.Int64 != folder.ID {
				s.logger.Info("Asset ownership change detected", "path", path, "old_folder", cached.ScanFolderID, "new_folder", folder.ID)
				err := s.db.UpdateAssetLocation(ctx, database.UpdateAssetLocationParams{
					ID:           cached.ID,
					FilePath:     path,
					ScanFolderID: sql.NullInt64{Int64: folder.ID, Valid: true},
					LastScanned:  time.Now(),
				})
				if err != nil {
					s.logger.Error("Failed to re-bind asset", "error", err)
				}
			}
			cachedTime := existingCache[path].LastModified
			diskTime := info.ModTime()
			if cached.IsDeleted {
				s.logger.Info("Restoring asset", "path", path)
				s.db.RestoreAsset(ctx, cached.ID)
				cached.IsDeleted = false
			}
			if !diskTime.Equal(cachedTime) {
				//TODO: Update logic, robimy checka nazwy pliku i innych danych
				// s.logger.Info("File modified, updating...", "path", path)
			}
			// Update procesu skanowania
			s.updateAndEmitTotal(total, d, totalToProcess)
			//Tak to skip
			return nil
		}
		var fileHash string
		hash, err := CalculateFileHash(path, s.config.MaxAllowHashFileSize)
		if err == nil {
			fileHash = hash
		} else {
			s.logger.Debug("Failed to calculate hash", "path", path, "error", err)
		}
		// --- SELF-HEALING & DUPLICATE DETECTION ---
		if fileHash != "" {
			existingAsset, err := s.db.GetAssetByHash(ctx, sql.NullString{String: fileHash, Valid: true})

			if err == nil {
				isMove := false

				if existingAsset.IsDeleted {
					// Przypadek 1: Plik był w koszu
					isMove = true
				} else {
					// Przypadek 2: Plik jest aktywny w bazie
					if _, statErr := os.Stat(existingAsset.FilePath); os.IsNotExist(statErr) {
						// Starego pliku nie ma na dysku
						isMove = true
					}
					// Jeśli istnieje -> To jest KOPIA.
				}

				if isMove {
					s.logger.Info("Self-healing: Asset moved", "old_path", existingAsset.FilePath, "new_path", path)
					err := s.db.UpdateAssetLocation(ctx, database.UpdateAssetLocationParams{
						ID:           existingAsset.ID,
						FilePath:     path,
						ScanFolderID: sql.NullInt64{Int64: int64(folder.ID), Valid: true},
						LastScanned:  time.Now(),
					})

					if err != nil {
						s.logger.Error("Failed to update moved asset", "error", err)
					} else {
						foundAssets[existingAsset.ID] = true
						s.updateAndEmitTotal(total, d, totalToProcess)
						return nil
					}
				} else {
					s.logger.Info("Duplicate detected (Copy)", "path", path, "original_id", existingAsset.ID)
				}
			}
		}
		thumbResult, err := s.thumbGen.Generate(ctx, path)
		if err != nil {
			s.logger.Error("Critical thumbnail generation failure", "path", path, "error", err)
		}
		// TODO: Add thumbnail generation logic
		hasValidDimensions := thumbResult.Metadata.Width > 0 && thumbResult.Metadata.Height > 0
		_, dbErr := s.db.CreateAsset(ctx, database.CreateAssetParams{
			ScanFolderID:    sql.NullInt64{Int64: int64(folder.ID), Valid: true},
			FileName:        d.Name(),
			FilePath:        path,
			FileType:        fileType,
			FileSize:        info.Size(),
			ThumbnailPath:   thumbResult.WebPath,
			FileHash:        sql.NullString{String: fileHash, Valid: fileHash != ""},
			ImageWidth:      sql.NullInt64{Int64: int64(thumbResult.Metadata.Width), Valid: hasValidDimensions},
			ImageHeight:     sql.NullInt64{Int64: int64(thumbResult.Metadata.Height), Valid: hasValidDimensions},
			DominantColor:   sql.NullString{String: string(thumbResult.Metadata.DominantColor), Valid: thumbResult.Metadata.DominantColor != ""},
			BitDepth:        sql.NullInt64{Int64: int64(thumbResult.Metadata.BitDepth), Valid: hasValidDimensions},
			HasAlphaChannel: sql.NullBool{Bool: thumbResult.Metadata.HasAlphaChannel, Valid: hasValidDimensions},
			LastModified:    info.ModTime(),
			LastScanned:     time.Now(),
		})
		if dbErr != nil {
			// TODO: Handle UNIQUE constraint violation
			s.logger.Debug("Skipping asset", "path", path, "reason", dbErr)
		} else {
			s.updateAndEmitTotal(total, d, totalToProcess)
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

	rows, err := s.db.ListAssetsForCache(ctx)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]CachedAsset, len(rows))

	for _, row := range rows {
		existing[row.FilePath] = CachedAsset{
			ID:           row.ID,
			LastModified: row.LastModified,
			IsDeleted:    row.IsDeleted,
			ScanFolderID: row.ScanFolderID,
		}
	}
	s.logger.Info("Loaded assets cache", "count", len(existing))
	return existing, nil

}

func (s *Scanner) updateAndEmitTotal(total *int, d fs.DirEntry, totalToProcess int) {
	*total++
	if *total%10 == 0 {
		runtime.EventsEmit(s.ctx, "scan_progress", map[string]any{
			"current":  *total,
			"total":    totalToProcess,
			"lastFile": d.Name(),
		})
	}
}
