package services

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	goRuntime "runtime"
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
	conn       *sql.DB
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
	Path         string
	Err          error
	NewAsset     *database.CreateAssetParams
	ExistingPath string
}

func (s *Scanner) Startup(ctx context.Context) {
	s.ctx = ctx
}

// NewScanner creates a new Scanner instance
func NewScanner(conn *sql.DB, db *database.Queries, thumbGen *ThumbnailGenerator) *Scanner {
	return &Scanner{
		conn:     conn,
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
	FilePath     string
}

func (s *Scanner) StartScan() error {
	if s.isScanning.Load() {
		return nil
	}
	s.isScanning.Store(true)
	scanCtx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	s.ctx = scanCtx
	jobs := make(chan ScanJob, 100)
	results := make(chan ScanResult, 100)

	var workersWg sync.WaitGroup
	numWorkers := goRuntime.NumCPU()
	s.logger.Info("Starting scanner", "workers", numWorkers)

	go func() {
		defer s.isScanning.Store(false)
		defer cancel()
		var folders []database.ScanFolder
		var totalToProcess int = 0
		var err error
		var foundOnDisk map[string]bool
		collectorDone := make(chan map[string]bool)

		existingAssets, err := s.loadExistingAssets(scanCtx)
		if err != nil {
			s.logger.Error("Failed to load existing assets. Aborting.", "error", err)
			return
		}

		folders, err = s.db.ListScanFolders(scanCtx)
		if err != nil {
			s.logger.Error("Failed to list folders", slog.String("error", err.Error()))
			return
		}
		s.logger.Info("Calculating total files...")
		for _, f := range folders {
			if scanCtx.Err() != nil {
				s.logger.Info("Scanner cancelled by user")
				break
			}
			totalToProcess += s.getAllFilesCount(f)
		}
		s.logger.Info("Total files to scan calculated", "count", totalToProcess)
		go func() {
			defer close(collectorDone)
			foundOnDisk = s.Collector(scanCtx, totalToProcess, results)
			collectorDone <- foundOnDisk
		}()
		for i := 0; i < numWorkers; i++ {
			workersWg.Add(1)
			go s.Worker(scanCtx, &workersWg, jobs, results)
		}
		// TODO:Implementacja pętli po folderach i WalkDir
		for _, f := range folders {
			if scanCtx.Err() != nil {
				break
			}
			err := s.scanDirectory(scanCtx, f, jobs)
			if err != nil {
				s.logger.Error("Failed to scan directory", slog.String("error", err.Error()))
			}
		}
		close(jobs)
		workersWg.Wait()
		close(results)
		foundOnDisk = <-collectorDone
		s.logger.Info("Scanner finished", "total", totalToProcess)
		runtime.EventsEmit(s.ctx, "scan_status", "idle")
		s.logger.Info("Scan finished. Starting Cleanup phase...",
			"db_cache_size", len(existingAssets),
			"found_on_disk", len(foundOnDisk))
		for path, cached := range existingAssets {
			if !foundOnDisk[cached.FilePath] {
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

// Worker to główna pętla przetwarzająca zadania skanowania.
// Działa jako orkiestrator: Hashowanie -> Lookup -> Delegacja decyzji.
func (s *Scanner) Worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan ScanJob, results chan<- ScanResult) {
	defer wg.Done()

	for job := range jobs {
		result := ScanResult{Path: job.Path}

		ext := filepath.Ext(job.Path)
		if !s.IsExtensionAllowed(ext) {
			continue
		}

		hash, err := CalculateFileHash(job.Path, s.config.MaxAllowHashFileSize)
		if err != nil {
			s.logger.Warn("Hashing failed", "path", job.Path, "error", err)
			result.Err = err
			results <- result
			continue
		}
		var exist database.Asset
		var lookupErr error

		if hash != "" {
			exist, lookupErr = s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
		} else {
			exist, lookupErr = s.db.GetAssetByPath(ctx, job.Path)
		}

		if lookupErr != nil && lookupErr != sql.ErrNoRows {
			s.logger.Error("DB Lookup Error", "path", job.Path, "error", lookupErr)
			result.Err = lookupErr
			results <- result
			continue
		}

		fileType := DetermineFileType(ext)

		if exist.ID > 0 {
			// Asset już znamy. Ale czy to ten sam? Czy zombie? Czy kopia?
			s.processExistingAsset(ctx, &result, exist, job, fileType, hash)
		} else {
			s.processNewAsset(ctx, &result, job, fileType, hash)
		}

		results <- result
	}
}

// processNewAsset obsługuje proste dodawanie nowego pliku.
func (s *Scanner) processNewAsset(ctx context.Context, result *ScanResult, job ScanJob, fileType, hash string) {
	if job.Entry == nil {
		return
	}

	s.logger.Debug("New asset detected", "path", job.Path)

	newAsset, err := s.generateAssetMetadata(ctx, job.Path, job.Entry, job.FolderId, fileType, hash)
	if err != nil {
		s.logger.Warn("Failed to generate metadata for new asset", "path", job.Path, "error", err)
		result.Err = err
	} else {
		result.NewAsset = &newAsset
	}
}

// processExistingAsset to serce logiki Self-Healing.
// Obsługuje: Zmiany nazw, Duplikaty, Przywracanie (Resurrection) i Refresh Metadanych.
func (s *Scanner) processExistingAsset(ctx context.Context, result *ScanResult, exist database.Asset, job ScanJob, fileType, hash string) {
	// SCENARIUSZ A: Ścieżki się zgadzają. To ten sam plik.
	if exist.FilePath == job.Path {
		// 1. Resurrection Check (Czy to Zombie?)
		// Plik jest na dysku, ale w bazie ma flagę is_deleted.
		// Sytuacja: Użytkownik przywrócił plik z kosza LUB odblokował rozszerzenie.
		if exist.IsDeleted {
			s.logger.Info("🧟 Resurrection: Restoring soft-deleted asset", "id", exist.ID, "path", job.Path)
			if err := s.db.RestoreAsset(ctx, exist.ID); err != nil {
				s.logger.Error("Failed to resurrect asset", "error", err)
			}
		}

		// 2. Integrity Check (Czy plik był edytowany?)
		// Sprawdzamy daty modyfikacji.
		info, err := job.Entry.Info()
		if err != nil {
			info, _ = os.Stat(job.Path)
		}

		if info != nil {
			dbTime := exist.LastModified.Unix()
			diskTime := info.ModTime().Unix()

			if dbTime != diskTime {
				s.logger.Info("📝 File Content Changed: Refreshing metadata", "path", job.Path)

				meta, err := s.generateAssetMetadata(ctx, job.Path, job.Entry, job.FolderId, fileType, hash)
				if err == nil {
					err = s.db.RefreshAssetTechnicalMetadata(ctx, database.RefreshAssetTechnicalMetadataParams{
						FileSize:        info.Size(),
						LastModified:    info.ModTime(),
						LastScanned:     time.Now(),
						ThumbnailPath:   meta.ThumbnailPath,
						ImageWidth:      meta.ImageWidth,
						ImageHeight:     meta.ImageHeight,
						DominantColor:   meta.DominantColor,
						BitDepth:        meta.BitDepth,
						HasAlphaChannel: meta.HasAlphaChannel,
						ID:              exist.ID,
					})
					if err != nil {
						s.logger.Error("Failed to save refreshed metadata", "error", err)
					}
				}
			}
		}

		// Oznaczamy jako istniejący dla Collectora (żeby Cleanup go nie usunął)
		result.ExistingPath = exist.FilePath
		return
	}

	// SCENARIUSZ B: Ścieżki się RÓŻNIĄ. (Move vs Copy)
	// Hash jest ten sam, ale plik jest w innym miejscu.

	// Sprawdzamy co się stało ze STARĄ lokalizacją (tą z bazy).
	_, statErr := os.Stat(exist.FilePath)
	oldFileMissing := os.IsNotExist(statErr)

	if oldFileMissing {
		// 1. MOVE (Przeniesienie / Zmiana nazwy)
		// Starego nie ma, nowy jest. To ten sam byt.
		s.logger.Info("🚚 Move/Rename Detected", "old_path", exist.FilePath, "new_path", job.Path)

		err := s.db.UpdateAssetLocation(ctx, database.UpdateAssetLocationParams{
			ID:           exist.ID,
			FilePath:     job.Path,
			LastScanned:  time.Now(),
			ScanFolderID: sql.NullInt64{Int64: job.FolderId, Valid: job.FolderId > 0},
		})

		if err != nil {
			s.logger.Error("Failed to update location for moved asset", "error", err)
			result.Err = err
		} else {
			// Cleanup ma w pamięci snapshot ze starą ścieżką. Musimy mu powiedzieć: "Spoko, ogarnąłem ten stary plik".
			result.ExistingPath = exist.FilePath
		}

	} else {
		// 2. COPY (Duplikat)
		// Stary jest, nowy też jest. To są dwa fizyczne byty.
		s.logger.Info("👯 Duplicate Detected (Copy)", "original_id", exist.ID, "new_copy_path", job.Path)

		// Traktujemy to jako zupełnie nowy asset.
		// TODO: W przyszłości dodamy linkowanie parent/child
		s.processNewAsset(ctx, result, job, fileType, hash)
	}
}
func (s *Scanner) Collector(ctx context.Context, totalToProcess int, results <-chan ScanResult) map[string]bool {
	const batchSize = 100
	const emitAfter = 50
	buff := make([]ScanResult, 0, batchSize)
	processed := make(map[string]bool)

	var totalProcessed int

	flush := func() {
		if len(buff) > 0 {
			s.logger.Info("Flushing batch to DB", "count", len(buff))
			err := s.InsertAssets(ctx, buff)
			if err != nil {
				s.logger.Error("Batch insert failed", "error", err)
			}
			buff = buff[:0]
		}
	}

	for result := range results {
		if result.Err != nil {
			s.logger.Error("Error scanning file", "path", result.Path, "error", result.Err)
		} else {
			s.logger.Debug("File scanned", "path", result.Path)
		}
		if result.NewAsset != nil {
			buff = append(buff, result)
		}

		// Jeśli Worker wykrył przeniesienie, podał nam starą ścieżkę w ExistingPath.
		// Musimy ją "odznaczyć", żeby Cleanup wiedział, że ten Asset ID (znany mu pod starą nazwą) przetrwał.
		if result.ExistingPath != "" {
			processed[result.ExistingPath] = true
		} else {
			processed[result.Path] = true
		}
		if len(buff) >= batchSize {
			flush()
		}
		s.updateAndEmitTotal(&totalProcessed, totalToProcess, emitAfter)
	}
	flush()
	return processed
}
func (s *Scanner) InsertAssets(ctx context.Context, buffer []ScanResult) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := s.db.WithTx(tx)
	for _, item := range buffer {
		if item.NewAsset != nil {
			_, err := qtx.CreateAsset(ctx, *item.NewAsset)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
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
func (s *Scanner) scanDirectory(scanCtx context.Context, folder database.ScanFolder, jobs chan<- ScanJob) error {
	if scanCtx.Err() != nil {
		return scanCtx.Err()
	}

	s.logger.Info("Scanning folder", "path", folder.Path)

	err := filepath.WalkDir(folder.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
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
		job := ScanJob{
			Path:     path,
			FolderId: folder.ID,
			Entry:    d,
		}
		select {
		case jobs <- job:
			// Sukces, idziemy dalej
		case <-scanCtx.Done():
			// Koniec zabawy, przerywamy WalkDir
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		s.logger.Error("WalkDir failed", "path", folder.Path, "error", err)
		return err
	}
	return nil
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
			FilePath:     row.FilePath,
		}
	}
	s.logger.Info("Loaded assets cache", "count", len(existing))
	return existing, nil

}

func (s *Scanner) updateAndEmitTotal(total *int, totalToProcess, emitAfter int) {
	*total++
	if *total%emitAfter == 0 {
		runtime.EventsEmit(s.ctx, "scan_progress", map[string]any{
			"current": *total,
			"total":   totalToProcess,
		})
	}
}
