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

// Scanner is a struct that keeps needed dependencies for scanning assets.
type Scanner struct {
	mu         sync.RWMutex
	db         *database.Queries
	logger     *slog.Logger
	cancelFunc context.CancelFunc
	config     ScannerConfig
	isScanning atomic.Bool
}

// NewScanner creates a new Scanner instance
func NewScanner(db *database.Queries) *Scanner {
	return &Scanner{
		db:     db,
		logger: slog.Default(),
		config: *NewScannerConfig(),
	}
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
		runtime.EventsEmit(wailsCtx, "scan_status", "scanning") // Needed for UI Scanner Status Update

		// Get active folders
		folders, err := s.db.ListScanFolders(scanCtx)
		if err != nil {
			s.logger.Error("Failed to list folders", slog.String("error", err.Error()))
		}

		totalProcessed := 0

		for _, f := range folders {
			// Check for the cancellation
			if scanCtx.Err() != nil {
				s.logger.Info("Scanner cancelled by user")
				break
			}
			s.scanDirectory(scanCtx, wailsCtx, f, &totalProcessed)
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
func (s *Scanner) scanDirectory(ctx context.Context, wailsCtx context.Context, folder database.ScanFolder, total *int) {
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
			return nil
		}

		// TODO: Add thumbnail generation logic

		info, _ := d.Info()

		_, dbErr := s.db.CreateAsset(ctx, database.CreateAssetParams{
			ScanFolderID:  sql.NullInt64{Int64: int64(folder.ID), Valid: true},
			FileName:      d.Name(),
			FilePath:      path,
			FileType:      "unknown", // TODO: Determine file type
			FileSize:      info.Size(),
			ThumbnailPath: "", // TODO: Generate thumbnail logic and path
			LastModified:  info.ModTime(),
			LastScanned:   time.Now(),
			// TODO: Add more fields as needed
		})
		if dbErr != nil {
			// TODO: Handle UNIQUE constraint violation
			s.logger.Debug("Skipping asset", "path", path, "reason", dbErr)
		} else {
			*total++
			if *total%10 == 0 {
				runtime.EventsEmit(wailsCtx, "scan_progress", map[string]any{
					"current":  *total,
					"lastFile": d.Name(),
				})
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Error("WalkDir failed", "path", folder.Path, "error", err)
	}
}
