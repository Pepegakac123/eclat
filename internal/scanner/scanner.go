package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/config"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Scanner struct {
	mu         sync.RWMutex
	db         database.Querier
	conn       *sql.DB
	logger     *slog.Logger
	cancelFunc context.CancelFunc
	config     *config.ScannerConfig // Zmiana typu
	isScanning atomic.Bool
	thumbGen   ThumbnailGenerator
	ctx        context.Context
	notifier   feedback.Notifier // Zmiana typu
}

type ScanJob struct {
	Path     string
	FolderId int64
	Entry    fs.DirEntry
}

type ScanResult struct {
	Path          string
	Err           error
	NewAsset      *database.CreateAssetParams
	ModifiedAsset *database.UpdateAssetFromScanParams
	ExistingPath  string
}

// fileInfoEntry to adapter, który pozwala użyć fs.FileInfo jako fs.DirEntry
type fileInfoEntry struct {
	info fs.FileInfo
}

func (e fileInfoEntry) Name() string               { return e.info.Name() }
func (e fileInfoEntry) IsDir() bool                { return e.info.IsDir() }
func (e fileInfoEntry) Type() fs.FileMode          { return e.info.Mode().Type() }
func (e fileInfoEntry) Info() (fs.FileInfo, error) { return e.info, nil }

func (s *Scanner) Startup(ctx context.Context) {
	s.ctx = ctx
}

func NewScanner(conn *sql.DB, db database.Querier, thumbGen ThumbnailGenerator, logger *slog.Logger, notifier feedback.Notifier) *Scanner {
	return &Scanner{
		conn:     conn,
		db:       db,
		logger:   logger,
		config:   config.NewScannerConfig(),
		thumbGen: thumbGen,
		notifier: notifier,
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
		s.notifier.SendScannerStatus(s.ctx, feedback.Scanning)
		s.notifier.SendScanProgress(s.ctx, 0, totalToProcess, "Initializing...")
		go func() {
			defer close(collectorDone)
			foundOnDisk = s.Collector(scanCtx, totalToProcess, results)
			collectorDone <- foundOnDisk
		}()
		for i := 0; i < numWorkers; i++ {
			workersWg.Add(1)
			go s.Worker(scanCtx, &workersWg, jobs, results)
		}
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
		s.notifier.SendScannerStatus(s.ctx, feedback.Idle)
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

func (s *Scanner) StopScan() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

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
			s.logger.Warn("Hashing skipped (likely too large)", "path", job.Path, "reason", err)
			hash = ""
		}
		var exist database.Asset
		var lookupErr error

		exist, lookupErr = s.db.GetAssetByPath(ctx, job.Path)
		if lookupErr == sql.ErrNoRows && hash != "" {
			exist, lookupErr = s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
		}

		if lookupErr != nil && lookupErr != sql.ErrNoRows {
			s.logger.Error("DB Lookup Error", "path", job.Path, "error", lookupErr)
			result.Err = lookupErr
			results <- result
			continue
		}

		fileType := DetermineFileType(ext)

		if exist.ID > 0 {
			s.processExistingAsset(ctx, &result, exist, job, fileType, hash)
		} else {
			s.processNewAsset(ctx, &result, job, fileType, hash)
		}

		results <- result
	}
}

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

func (s *Scanner) processExistingAsset(ctx context.Context, result *ScanResult, exist database.Asset, job ScanJob, fileType, hash string) {
	// Sprawdzamy czy ścieżka się zgadza (To ten sam plik na dysku co w bazie)
	if exist.FilePath == job.Path {

		// Scenariusz A: RESURRECTION (Przywracanie z kosza)
		// Jeśli plik był oznaczony jako usunięty, ale go znaleźliśmy -> ożywiamy go.
		var isResurrected bool
		if exist.IsDeleted {
			s.logger.Info("🧟 Resurrection: Restoring soft-deleted asset", "id", exist.ID, "path", job.Path)
			isResurrected = true
		}

		info, err := job.Entry.Info()
		if err != nil {
			info, _ = os.Stat(job.Path)
		}

		// Scenariusz B: CONTENT REFRESH (Zmiana zawartości)
		if info != nil {
			dbTime := exist.LastModified.Unix()
			diskTime := info.ModTime().Unix()

			// Odświeżamy jeśli: (czas się zmienił) LUB (był martwy i wstaje z grobu)
			if dbTime != diskTime || isResurrected {
				s.logger.Info("📝 File Content Changed or Resurrected: Refreshing metadata", "path", job.Path)

				meta, err := s.generateAssetMetadata(ctx, job.Path, job.Entry, job.FolderId, fileType, hash)
				if err == nil {
					// Tworzymy PACZKĘ aktualizacyjną
					modifiedAsset := &database.UpdateAssetFromScanParams{
						ID: exist.ID,
						// WAŻNE: Tu "ożywiamy" plik (IsDeleted = false)
						IsDeleted:   sql.NullBool{Bool: false, Valid: true},
						LastScanned: sql.NullTime{Time: time.Now(), Valid: true},

						// Metadane techniczne
						FileSize:     sql.NullInt64{Int64: meta.FileSize, Valid: true},
						LastModified: sql.NullTime{Time: meta.LastModified, Valid: true},
						FileHash:     meta.FileHash,

						// Metadane wizualne
						ThumbnailPath:   sql.NullString{String: meta.ThumbnailPath, Valid: true},
						ImageWidth:      meta.ImageWidth,
						ImageHeight:     meta.ImageHeight,
						DominantColor:   meta.DominantColor,
						BitDepth:        meta.BitDepth,
						HasAlphaChannel: meta.HasAlphaChannel,
					}
					result.ModifiedAsset = modifiedAsset
				}
			} else if isResurrected {
				// Jeśli tylko ożywiamy, ale treść się nie zmieniła, wysyłamy minimalny update
				result.ModifiedAsset = &database.UpdateAssetFromScanParams{
					ID:          exist.ID,
					IsDeleted:   sql.NullBool{Bool: false, Valid: true},
					LastScanned: sql.NullTime{Time: time.Now(), Valid: true},
				}
			}
		}
		result.ExistingPath = exist.FilePath
		return
	}

	// 2. Jeśli ścieżki się różnią, ale znaleźliśmy plik w bazie (po hashu?)
	// To oznacza Move/Rename lub Kopię.

	_, statErr := os.Stat(exist.FilePath)
	oldFileMissing := os.IsNotExist(statErr)

	if oldFileMissing {
		// Scenariusz C: MOVE / RENAME
		s.logger.Info("🚚 Move/Rename Detected", "old_path", exist.FilePath, "new_path", job.Path)

		result.ModifiedAsset = &database.UpdateAssetFromScanParams{
			ID:           exist.ID,
			FilePath:     sql.NullString{String: job.Path, Valid: true},
			ScanFolderID: sql.NullInt64{Int64: job.FolderId, Valid: true},
			IsDeleted:    sql.NullBool{Bool: false, Valid: true}, // Upewniamy się, że żyje
			LastScanned:  sql.NullTime{Time: time.Now(), Valid: true},
		}
		result.ExistingPath = exist.FilePath

	} else {
		// Scenariusz D: DUPLICATE (Copy)
		// Tutaj tworzymy nowy asset, więc processNewAsset jest OK.
		s.logger.Info("👯 Duplicate Detected (Copy)", "original_id", exist.ID, "new_copy_path", job.Path)
		s.processNewAsset(ctx, result, job, fileType, hash)
	}
}

func (s *Scanner) Collector(ctx context.Context, totalToProcess int, results <-chan ScanResult) map[string]bool {
	const batchSize = 100
	const emitAfter = 30 // UI update frequency
	buff := make([]ScanResult, 0, batchSize)
	processed := make(map[string]bool)

	var totalProcessed int

	flush := func() {
		if len(buff) > 0 {
			// s.logger.Info("Flushing batch to DB", "count", len(buff))
			err := s.ApplyBatch(ctx, buff)
			if err != nil {
				s.logger.Error("Batch operation failed", "error", err)
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
		// Dodajemy do bufora jeśli jest co zapisać (Nowy LUB Zmodyfikowany)
		if result.NewAsset != nil || result.ModifiedAsset != nil {
			buff = append(buff, result)
		}

		// Logika cache'owania ścieżek
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

func (s *Scanner) ApplyBatch(ctx context.Context, buffer []ScanResult) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := database.New(tx)
	if len(buffer) == 0 {
		return nil
	}

	for _, item := range buffer {
		// 1. INSERT
		if item.NewAsset != nil {
			_, err := qtx.CreateAsset(ctx, *item.NewAsset)
			if err != nil {
				s.logger.Error("Failed to insert asset", "path", item.Path, "error", err)
				continue
			}
		}

		// 2. UPDATE (Patch)
		if item.ModifiedAsset != nil {
			_, err := qtx.UpdateAssetFromScan(ctx, *item.ModifiedAsset)
			if err != nil {
				s.logger.Error("Failed to update asset", "path", item.Path, "error", err)
				continue
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	notifyCtx := ctx
	if s.ctx != nil {
		notifyCtx = s.ctx
	}

	s.notifier.EmitAssetsChanged(notifyCtx)

	return nil
}

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
		case <-scanCtx.Done():
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

// ScanFile przetwarza pojedynczy plik zgłoszony przez Watchera
func (s *Scanner) ScanFile(ctx context.Context, path string) error {
	s.logger.Info("⚡ Live Scan triggered", "path", path)

	ext := filepath.Ext(path)
	if !s.IsExtensionAllowed(ext) {
		return nil
	}

	info, statErr := os.Stat(path)
	fileExists := statErr == nil
	fileMissing := os.IsNotExist(statErr)
	asset, dbErr := s.db.GetAssetByPath(ctx, path)
	isKnown := dbErr == nil && asset.ID > 0
	// --- SCENARIUSZ 1: USUWANIE (Live Delete) ---
	if fileMissing {
		if isKnown && !asset.IsDeleted {
			s.logger.Info("🗑️ Soft Deleting asset", "path", path)
			return s.db.SoftDeleteAsset(ctx, asset.ID)
		}
		return nil
	}
	if !fileExists {
		s.logger.Warn("File access error during live scan", "path", path, "error", statErr)
		return nil
	}
	hash, err := CalculateFileHash(path, s.config.MaxAllowHashFileSize)
	if err != nil {
		s.logger.Warn("Hashing failed (likely locked or too big)", "error", err)
	}

	result := ScanResult{Path: path}

	folderID, err := s.resolveFolderID(ctx, path)
	if err != nil {
		s.logger.Warn("Could not resolve ScanFolder ID for file", "path", path)
		// Kontynuujemy z folderID = 0 (baza obsłuży to jako NULL)
	}

	// 6. Przygotowanie Joba z Adapterem!
	fileType := DetermineFileType(ext)
	job := ScanJob{
		Path:     path,
		FolderId: folderID,
		Entry:    fileInfoEntry{info: info}, // <--- TU JEST MAGIA ADAPTERA
	}

	if isKnown {
		// Update / Move / Resurrect
		s.processExistingAsset(ctx, &result, asset, job, fileType, hash)
	} else {
		// Insert
		// Próbujemy jeszcze znaleźć po hashu (może to rename/move, którego nie złapaliśmy po ścieżce)
		if hash != "" {
			existingByHash, err := s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
			if err == nil && existingByHash.ID > 0 {
				s.processExistingAsset(ctx, &result, existingByHash, job, fileType, hash)
			} else {
				s.processNewAsset(ctx, &result, job, fileType, hash)
			}
		} else {
			s.processNewAsset(ctx, &result, job, fileType, hash)
		}
	}

	if result.NewAsset != nil || result.ModifiedAsset != nil {
		return s.ApplyBatch(ctx, []ScanResult{result})
	}

	return nil
}
func (s *Scanner) resolveFolderID(ctx context.Context, filePath string) (int64, error) {
	folders, err := s.db.ListScanFolders(ctx)
	if err != nil {
		return 0, err
	}

	var bestMatchID int64
	longestPrefixLen := 0

	cleanPath := filepath.Clean(filePath)

	for _, f := range folders {
		folderPath := filepath.Clean(f.Path)
		rel, err := filepath.Rel(folderPath, cleanPath)
		if err == nil && !strings.HasPrefix(rel, "..") {
			if len(folderPath) > longestPrefixLen {
				longestPrefixLen = len(folderPath)
				bestMatchID = f.ID
			}
		}
	}

	if bestMatchID == 0 {
		return 0, fmt.Errorf("no matching scan folder found")
	}

	return bestMatchID, nil
}
func (s *Scanner) ListenToWatcher(events <-chan string) {
	s.logger.Info("🔌 Scanner connected to Watcher events")
	for path := range events {
		ctx := context.Background()
		if s.ctx != nil {
			ctx = s.ctx
		}

		err := s.ScanFile(ctx, path)
		if err != nil {
			s.logger.Error("Live scan failed", "path", path, "error", err)
		} else {
			s.notifier.SendScannerStatus(s.ctx, feedback.Scanning)
		}
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
			FilePath:     row.FilePath,
		}
	}
	s.logger.Info("Loaded assets cache", "count", len(existing))
	return existing, nil
}

func (s *Scanner) updateAndEmitTotal(total *int, totalToProcess, emitAfter int) {
	*total++
	if *total%emitAfter == 0 {
		s.notifier.SendScanProgress(s.ctx, *total, totalToProcess, "")
	}
}
