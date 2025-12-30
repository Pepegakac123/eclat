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

	"github.com/google/uuid"
)

// Scanner is responsible for scanning the file system for assets,
// generating thumbnails, and synchronizing the state with the database.
type Scanner struct {
	mu         sync.RWMutex
	db         database.Querier
	conn       *sql.DB
	logger     *slog.Logger
	cancelFunc context.CancelFunc
	config     *config.ScannerConfig
	isScanning atomic.Bool
	thumbGen   ThumbnailGenerator
	ctx        context.Context
	notifier   feedback.Notifier

	// sessionCache and sessionMu are used for intra-scan duplicate detection.
	// Since database writes are batched, a worker might process a duplicate file
	// before its "original" is committed. This cache allows workers to share
	// newly generated GroupIDs for identical hashes instantly.
	sessionMu    sync.Mutex
	sessionCache map[string]string
}

// ScanJob represents a single file scanning task to be processed by a worker.
type ScanJob struct {
	Path     string
	FolderId int64
	Entry    fs.DirEntry
}

// ScanResult contains the outcome of a ScanJob, including any errors
// or database parameters for creating or updating an asset.
type ScanResult struct {
	Path          string
	Err           error
	NewAsset      *database.CreateAssetParams
	ModifiedAsset *database.UpdateAssetFromScanParams
	ExistingPath  string
}

// fileInfoEntry is an adapter that allows fs.FileInfo to satisfy the fs.DirEntry interface.
type fileInfoEntry struct {
	info fs.FileInfo
}

func (e fileInfoEntry) Name() string               { return e.info.Name() }
func (e fileInfoEntry) IsDir() bool                { return e.info.IsDir() }
func (e fileInfoEntry) Type() fs.FileMode          { return e.info.Mode().Type() }
func (e fileInfoEntry) Info() (fs.FileInfo, error) { return e.info, nil }

// Startup initializes the scanner with the provided context.
// It is called when the application starts.
func (s *Scanner) Startup(ctx context.Context) {
	s.ctx = ctx
}

// Shutdown gracefully stops any ongoing scan and performs necessary cleanup.
func (s *Scanner) Shutdown() {
	s.logger.Info("🛑 Stopping Scanner...")
	s.StopScan()
}

// NewScanner creates a new instance of Scanner.
// It requires a database connection, a thumbnail generator, a logger, and configuration.
func NewScanner(conn *sql.DB, db database.Querier, thumbGen ThumbnailGenerator, logger *slog.Logger, notifier feedback.Notifier, cfg *config.ScannerConfig) *Scanner {
	return &Scanner{
		conn:         conn,
		db:           db,
		logger:       logger,
		config:       cfg,
		thumbGen:     thumbGen,
		notifier:     notifier,
		sessionCache: make(map[string]string),
	}
}

// CachedAsset represents a minimal subset of asset data needed for synchronization
// checks during the scanning process.
type CachedAsset struct {
	ID           int64
	LastModified time.Time
	IsDeleted    bool
	ScanFolderID sql.NullInt64
	FilePath     string
}

// StartScan initiates the background scanning process.
// It checks all configured folders for new, modified, or deleted files.
// The scan runs asynchronously and utilizes a worker pool for parallel processing.
// If a scan is already in progress, this method returns nil immediately.
func (s *Scanner) StartScan() error {
	if s.isScanning.Load() {
		return nil
	}
	s.isScanning.Store(true)
	scanCtx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Initialize session cache for this specific scan run.
	s.sessionMu.Lock()
	s.sessionCache = make(map[string]string)
	s.sessionMu.Unlock()

	// Channels for the pipeline:
	// WalkDir -> jobs -> Workers -> results -> Collector
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

		// 1. Snapshot State: Load existing assets to detect deletions later.
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

		// 2. Preparation: Count files to show accurate progress bar.
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

		// 3. Start Collector: Runs in background to batch DB writes.
		go func() {
			defer close(collectorDone)
			// Collector returns the set of all file paths actually found on disk.
			foundOnDisk = s.Collector(scanCtx, totalToProcess, results)
			collectorDone <- foundOnDisk
		}()

		// 4. Start Workers: Parallel processing of files.
		for i := 0; i < numWorkers; i++ {
			workersWg.Add(1)
			go s.Worker(scanCtx, &workersWg, jobs, results)
		}

		// 5. Feed Workers: Walk directories and send jobs.
		for _, f := range folders {
			if scanCtx.Err() != nil {
				break
			}
			err := s.scanDirectory(scanCtx, f, jobs)
			if err != nil {
				s.logger.Error("Failed to scan directory", slog.String("error", err.Error()))
			}
		}

		// 6. Teardown: Close channels and wait for completion.
		close(jobs)
		workersWg.Wait() // Wait for workers to finish processing
		close(results)   // Close results to let Collector finish
		foundOnDisk = <-collectorDone

		s.logger.Info("Scanner finished", "total", totalToProcess)
		s.notifier.SendScannerStatus(s.ctx, feedback.Idle)

		// 7. Mark-and-Sweep Cleanup:
		// Any asset in DB (existingAssets) that was NOT found in the current scan (foundOnDisk)
		// must be marked as deleted.
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

// StopScan signals the current scan to cancel and stop.
func (s *Scanner) StopScan() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

// Worker consumes ScanJobs from the jobs channel, processes them, and sends the results to the results channel.
// It handles extension filtering, hashing, database lookups, and decides whether to create a new asset or update an existing one.
func (s *Scanner) Worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan ScanJob, results chan<- ScanResult) {
	defer wg.Done()

	for job := range jobs {
		result := ScanResult{Path: job.Path}

		ext := strings.ToLower(filepath.Ext(job.Path))
		if !s.IsExtensionAllowed(ext) {
			s.logger.Debug("Skipping file: extension not allowed", "path", job.Path, "ext", ext)
			continue
		}

		s.logger.Debug("Processing file", "path", job.Path, "ext", ext)

		// Calculate hash to detect duplicates or content changes.
		// Large files might be skipped (max hash size config).
		hash, err := CalculateFileHash(job.Path, s.config.GetMaxHashFileSize())
		if err != nil {
			s.logger.Warn("Hashing skipped (likely too large or error)", "path", job.Path, "error", err)
			hash = ""
		}
		var exist database.Asset
		var lookupErr error

		// Lookup Logic:
		// 1. Check if we know this specific path.
		exist, lookupErr = s.db.GetAssetByPath(ctx, job.Path)

		// 2. If path unknown but we have a hash, check if we know the content (Renames/Moves).
		if lookupErr == sql.ErrNoRows && hash != "" {
			exist, lookupErr = s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
			if lookupErr == nil {
				s.logger.Debug("Found existing asset by hash (Rename/Move detected)", "path", job.Path, "old_path", exist.FilePath)
			}
		}

		if lookupErr != nil && lookupErr != sql.ErrNoRows {
			s.logger.Error("DB Lookup Error", "path", job.Path, "error", lookupErr)
			result.Err = lookupErr
			results <- result
			continue
		}

		fileType := DetermineFileType(ext)
		s.logger.Debug("Determined file type", "path", job.Path, "type", fileType)

		if exist.ID > 0 {
			s.processExistingAsset(ctx, &result, exist, job, fileType, hash)
		} else {
			s.processNewAsset(ctx, &result, job, fileType, hash)
		}

		results <- result
	}
}

// processNewAsset handles the logic for a file that is not yet known in the database.
// It attempts to detect duplicates via hash or heuristics before creating a new asset record.
func (s *Scanner) processNewAsset(ctx context.Context, result *ScanResult, job ScanJob, fileType, hash string) {
	if job.Entry == nil {
		return
	}

	s.logger.Debug("New asset detected", "path", job.Path)

	targetGroupID := ""
	foundMatch := false

	// === PLAN A: EXACT MATCH (HASH) - DB LOOKUP ===
	// Check if this file is a duplicate of something already in the DB.
	if hash != "" {
		existingDuplicate, err := s.db.GetAssetByHash(ctx, sql.NullString{String: hash, Valid: true})
		if err == nil {
			s.logger.Info("🔗 Exact Duplicate found (DB)",
				"new_path", job.Path,
				"group_id", existingDuplicate.GroupID)
			targetGroupID = existingDuplicate.GroupID
			foundMatch = true
		}
	}

	// === PLAN A.5: EXACT MATCH (HASH) - SESSION CACHE LOOKUP ===
	// Check if another worker found this same file content in this current scan session.
	// This handles the case where we have 2 copies of a new file; one worker processes Copy A,
	// assigns a new GroupID, and caches it. The second worker processing Copy B hits this cache.
	if !foundMatch && hash != "" {
		s.sessionMu.Lock()
		if cachedGroupID, ok := s.sessionCache[hash]; ok {
			s.logger.Info("🔗 Exact Duplicate found (Session Cache)",
				"new_path", job.Path,
				"group_id", cachedGroupID)
			targetGroupID = cachedGroupID
			foundMatch = true
		}
		s.sessionMu.Unlock()
	}

	// === PLAN B: HEURISTIC MATCH (NAME) ===
	// If hashes don't match (or file too big to hash), try to guess relationships based on filenames.
	// Useful for grouping texture sets (e.g., "Wood_Color.png", "Wood_Normal.png").
	if !foundMatch {
		matchedGroupID, found := s.TryHeuristicMatch(ctx, job.FolderId, job.Entry.Name())
		if found {
			s.logger.Info("🧠 Heuristic Match found (Name)",
				"new_path", job.Path,
				"group_id", matchedGroupID)
			targetGroupID = matchedGroupID
			foundMatch = true
		}
	}

	// If no match found, this is a completely unique new asset.
	if !foundMatch {
		if hash != "" {
			s.sessionMu.Lock()
			// Double-check locking pattern: checking if another worker updated the cache
			// while we were performing heuristic checks.
			if cached, ok := s.sessionCache[hash]; ok {
				targetGroupID = cached
				foundMatch = true
			} else {
				targetGroupID = uuid.New().String()
				s.sessionCache[hash] = targetGroupID
			}
			s.sessionMu.Unlock()
		} else {
			targetGroupID = uuid.New().String()
		}
	}

	newAsset, err := s.generateAssetMetadata(ctx, job.Path, job.Entry, job.FolderId, fileType, hash, targetGroupID)
	if err != nil {
		s.logger.Error("Critical failure generating metadata for new asset", "path", job.Path, "error", err)
		result.Err = err
		return
	}
	result.NewAsset = &newAsset
}

// processExistingAsset handles the logic for a file that is already present in the database.
// It checks for modifications, content changes, moves, or renames.
func (s *Scanner) processExistingAsset(ctx context.Context, result *ScanResult, exist database.Asset, job ScanJob, fileType, hash string) {
	// Case 1: The path matches exactly.
	// This is an update to an existing file (or a resurrection if it was deleted).
	if exist.FilePath == job.Path {
		var isResurrected bool
		if exist.IsDeleted {
			s.logger.Info("🧟 Resurrection: Restoring soft-deleted asset", "id", exist.ID, "path", job.Path)
			isResurrected = true
		}

		info, err := job.Entry.Info()
		if err != nil {
			info, _ = os.Stat(job.Path)
		}

		if info != nil {
			dbTime := exist.LastModified.Unix()
			diskTime := info.ModTime().Unix()

			// If timestamp changed or we are resurrecting, we regenerate metadata.
			if dbTime != diskTime || isResurrected {
				s.logger.Info("📝 File Content Changed or Resurrected: Refreshing metadata", "path", job.Path)

				meta, err := s.generateAssetMetadata(ctx, job.Path, job.Entry, job.FolderId, fileType, hash, exist.GroupID)
				if err == nil {
					modifiedAsset := &database.UpdateAssetFromScanParams{
						ID:              exist.ID,
						IsDeleted:       sql.NullBool{Bool: false, Valid: true},
						LastScanned:     sql.NullTime{Time: time.Now(), Valid: true},
						FileSize:        sql.NullInt64{Int64: meta.FileSize, Valid: true},
						LastModified:    sql.NullTime{Time: meta.LastModified, Valid: true},
						FileHash:        meta.FileHash,
						ThumbnailPath:   sql.NullString{String: meta.ThumbnailPath, Valid: true},
						ImageWidth:      meta.ImageWidth,
						ImageHeight:     meta.ImageHeight,
						DominantColor:   meta.DominantColor,
						BitDepth:        meta.BitDepth,
						HasAlphaChannel: meta.HasAlphaChannel,
					}
					result.ModifiedAsset = modifiedAsset
				} else {
					s.logger.Error("Failed to regenerate metadata for existing asset", "path", job.Path, "error", err)
					// Fallback: just un-delete if it was resurrected
					if isResurrected {
						result.ModifiedAsset = &database.UpdateAssetFromScanParams{
							ID:          exist.ID,
							IsDeleted:   sql.NullBool{Bool: false, Valid: true},
							LastScanned: sql.NullTime{Time: time.Now(), Valid: true},
						}
					}
				}
			} else if isResurrected {
				// Resurrected but content seems same (rare but possible if logic changes), just un-delete.
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

	// Case 2: The path is DIFFERENT, but we found the record via Hash.
	// This means it's either a MOVE/RENAME or a COPY (Duplicate).
	_, statErr := os.Stat(exist.FilePath)
	oldFileMissing := os.IsNotExist(statErr)

	if oldFileMissing {
		// The old file is gone, so this must be a Rename/Move.
		// We update the existing record with the new path.
		s.logger.Info("🚚 Move/Rename Detected", "old_path", exist.FilePath, "new_path", job.Path)
		result.ModifiedAsset = &database.UpdateAssetFromScanParams{
			ID:           exist.ID,
			FilePath:     sql.NullString{String: job.Path, Valid: true},
			ScanFolderID: sql.NullInt64{Int64: job.FolderId, Valid: true},
			IsDeleted:    sql.NullBool{Bool: false, Valid: true},
			LastScanned:  sql.NullTime{Time: time.Now(), Valid: true},
		}
		result.ExistingPath = exist.FilePath // Mark old path as processed/handled
	} else {
		// The old file still exists. This is a Copy (Duplicate).
		// We treat it as a new asset, but processNewAsset will handle linking it via GroupID.
		s.logger.Info("👯 Duplicate Detected (Copy)", "original_id", exist.ID, "new_copy_path", job.Path)
		s.processNewAsset(ctx, result, job, fileType, hash)
	}
}

// Collector collects results from workers, batches them, and commits them to the database.
// It returns a map of all paths found on disk during the scan.
func (s *Scanner) Collector(ctx context.Context, totalToProcess int, results <-chan ScanResult) map[string]bool {
	const batchSize = 100
	const emitAfter = 30
	buff := make([]ScanResult, 0, batchSize)
	processed := make(map[string]bool)

	var totalProcessed int

	// Helper to flush current buffer to DB
	flush := func() {
		if len(buff) > 0 {
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
		if result.NewAsset != nil || result.ModifiedAsset != nil {
			buff = append(buff, result)
		}
		// Track which paths we've seen to enable cleanup of missing files later.
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
	flush() // Flush remaining items
	return processed
}

// ApplyBatch executes a batch of database insertions and updates within a single transaction.
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
		if item.NewAsset != nil {
			_, err := qtx.CreateAsset(ctx, *item.NewAsset)
			if err != nil {
				s.logger.Error("Failed to insert asset", "path", item.Path, "error", err)
				continue
			}
		}
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

	s.logger.Info("Successfully committed batch to DB", "count", len(buffer))

	notifyCtx := ctx
	if s.ctx != nil {
		notifyCtx = s.ctx
	}
	s.notifier.EmitAssetsChanged(notifyCtx)
	return nil
}

// generateAssetMetadata creates the necessary metadata parameters for a new or updated asset.
// It generates a thumbnail and extracts file information.
func (s *Scanner) generateAssetMetadata(ctx context.Context, path string, entry fs.DirEntry, folderId int64, filetype string, hash string, targetGroupID string) (database.CreateAssetParams, error) {
	thumb, err := s.thumbGen.Generate(ctx, path)
	if err != nil {
		s.logger.Warn("Failed to generate thumbnail, proceeding without it", "path", path, "error", err)
		// We don't return error here, because we still want to add the asset to the DB
		thumb = ThumbnailResult{
			WebPath: "/placeholders/generic_placeholder.webp",
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		s.logger.Debug("Failed to get file info", "path", path, "error", err)
		return database.CreateAssetParams{}, err
	}

	hasValidDimensions := thumb.Metadata.Width > 0 && thumb.Metadata.Height > 0
	newAsset := database.CreateAssetParams{
		ScanFolderID:    sql.NullInt64{Int64: folderId, Valid: true},
		GroupID:         targetGroupID,
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

// scanDirectory recursively walks a directory, creating ScanJobs for allowed files.
// It skips subdirectories that are not explicitly part of the scan if necessary (though current logic walks all).
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

// ScanFile performs a targeted scan of a single file.
// This is typically called by the file watcher when a file event occurs.
// It handles new files, modifications, and deletions.
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

	hash, err := CalculateFileHash(path, s.config.GetMaxHashFileSize())
	if err != nil {
		s.logger.Warn("Hashing failed (likely locked or too big)", "error", err)
	}

	result := ScanResult{Path: path}

	folderID, err := s.resolveFolderID(ctx, path)
	if err != nil {
		s.logger.Warn("Could not resolve ScanFolder ID for file", "path", path)
	}

	fileType := DetermineFileType(ext)
	job := ScanJob{
		Path:     path,
		FolderId: folderID,
		Entry:    fileInfoEntry{info: info},
	}
	if isKnown {
		// Scenariusz 1: Plik jest w bazie pod tą ścieżką.
		s.processExistingAsset(ctx, &result, asset, job, fileType, hash)
	} else {
		// Scenariusz 2: Plik nie jest w bazie pod tą ścieżką.
		s.processNewAsset(ctx, &result, job, fileType, hash)
	}

	if result.NewAsset != nil || result.ModifiedAsset != nil {
		return s.ApplyBatch(ctx, []ScanResult{result})
	}

	return nil
}

// resolveFolderID matches a file path to the specific ScanFolder ID it belongs to.
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

// ListenToWatcher consumes file system events from the provided channel
// and triggers `ScanFile` for each event.
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

// getAllFilesCount counts the number of allowed files in a scan folder for progress reporting.
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

// loadExistingAssets fetches all known assets from the database into a memory map
// to facilitate fast existence checks during the scan.
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

// updateAndEmitTotal emits a progress event via the notifier every `emitAfter` items.
func (s *Scanner) updateAndEmitTotal(total *int, totalToProcess, emitAfter int) {
	*total++
	if *total%emitAfter == 0 {
		s.notifier.SendScanProgress(s.ctx, *total, totalToProcess, "")
	}
}
