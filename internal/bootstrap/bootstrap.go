package bootstrap

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"eclat/internal/app"
	"eclat/internal/config"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"eclat/internal/update"
	"eclat/internal/watcher"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // SQLite driver
)

// syncedWriter ensures that every write is followed by a Sync() call on Windows,
// preventing empty/broken log files when viewed while the app is running.
type syncedWriter struct {
	f *os.File
}

func (w *syncedWriter) Write(p []byte) (n int, err error) {
	n, err = w.f.Write(p)
	if err == nil && runtime.GOOS == "windows" {
		_ = w.f.Sync()
	}
	return
}

// Dependencies holds all the initialized services and resources required by the application.
type Dependencies struct {
	DB                 *sql.DB
	LogFile            *os.File
	Logger             *slog.Logger
	App                *app.App
	AssetService       *app.AssetService
	MaterialSetService *app.MaterialSetService
	TagService         *app.TagService
	ScannerService     *scanner.Scanner
	SettingsService    *settings.SettingsService
	WatcherService     *watcher.Service
	UpdateService      *update.UpdateService
	ThumbnailsDir      string
}

// Close closes all open resources like the database and log file.
func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
	if d.LogFile != nil {
		d.LogFile.Close()
	}
}

// Initialize performs the startup sequence: configuring directories, logger, database, and services.
func Initialize(migrations embed.FS) (*Dependencies, error) {
	// 1. Setup Directories
	// Note: We use UserCacheDir for the database and logs as it's standard for application data 
	// that doesn't need to be backed up by the system, and it ensures data persists across updates
	// because it's stored in the user profile, not the application installation directory.
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get cache dir: %w", err)
	}

	appCachePath := filepath.Join(userCacheDir, "eclat")
	dbFolder := filepath.Join(appCachePath, "db")
	thumbsFolder := filepath.Join(appCachePath, "thumbnails")
	logsFolder := filepath.Join(appCachePath, "logs")

	dirsToCreate := []string{appCachePath, dbFolder, thumbsFolder, logsFolder}
	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}

	// 2. Setup Logger
	currentTime := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logsFolder, fmt.Sprintf("app-%s.log", currentTime))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, &syncedWriter{f: logFile})
	programLogger := slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(programLogger)

	// 3. Setup Database
	dbPath := filepath.Join(dbFolder, "assets.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	// 4. Run Migrations
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}
	// Goose expects the path relative to the embed root
	if err := goose.Up(db, "sql/schema"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// 5. Initialize Core Logic & Configuration
	queries := database.New(db)
	sharedConfig := config.NewScannerConfig()
	ctx := context.Background()

	programLogger.Info("Loading system settings...")
	storedExtsJSON, err := queries.GetSystemSetting(ctx, "allowed_extensions")
	if err == nil && storedExtsJSON != "" {
		var storedExts []string
		if err := json.Unmarshal([]byte(storedExtsJSON), &storedExts); err == nil {
			if len(storedExts) > 0 {
				programLogger.Info("✅ Restored extensions from DB", "count", len(storedExts))
				sharedConfig.SetAllowedExtensions(storedExts)
			} else {
				programLogger.Info("⚠️ DB extensions list is empty, using defaults")
			}
		} else {
			programLogger.Error("❌ Failed to unmarshal settings from DB, using defaults", "error", err)
		}
	} else {
		programLogger.Info("ℹ️ No custom settings found in DB, using defaults")
	}

	// 6. Initialize Services
	notifier := feedback.NewNotifier()
	diskThumbGen := scanner.NewDiskThumbnailGenerator(thumbsFolder, programLogger)

	scannerService := scanner.NewScanner(db, queries, diskThumbGen, programLogger, notifier, sharedConfig)

	watcherService, err := watcher.NewService(queries, programLogger, sharedConfig)
	if err != nil {
		_ = db.Close()
		log.Fatal("Failed to create watcher:", err)
	}

	settingsService := settings.NewSettingsService(queries, programLogger, notifier, watcherService, sharedConfig)
	assetService := app.NewAssetService(queries, db, programLogger, notifier, thumbsFolder)
	materialSetService := app.NewMaterialSetService(queries, programLogger, diskThumbGen)
	tagService := app.NewTagService(queries, programLogger)
	updateService := update.NewUpdateService(programLogger)

	myApp := app.NewApp(queries, programLogger, assetService, materialSetService, tagService, scannerService, settingsService, watcherService, updateService)

	// 7. Cleanup Old Data (Logs and Soft-Deleted Assets older than 7 days)
	cleanupOldData(logsFolder, queries, programLogger)

	return &Dependencies{
		DB:                 db,
		LogFile:            logFile,
		Logger:             programLogger,
		App:                myApp,
		AssetService:       assetService,
		MaterialSetService: materialSetService,
		TagService:         tagService,
		ScannerService:     scannerService,
		SettingsService:    settingsService,
		WatcherService:     watcherService,
		UpdateService:      updateService,
		ThumbnailsDir:      thumbsFolder,
	}, nil
}

// cleanupOldData removes log files and database entries that are older than 7 days.
func cleanupOldData(logsFolder string, queries *database.Queries, logger *slog.Logger) {
	// 1. Cleanup Logs older than 7 days
	files, err := os.ReadDir(logsFolder)
	if err == nil {
		now := time.Now()
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, err := file.Info()
			if err != nil {
				continue
			}
			// Delete if older than 7 days
			if now.Sub(info.ModTime()) > 7*24*time.Hour {
				err := os.Remove(filepath.Join(logsFolder, file.Name()))
				if err == nil {
					logger.Info("🗑️ Deleted old log file", "name", file.Name())
				}
			}
		}
	}

	// 2. Cleanup Old Deleted Assets (Database entries)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := queries.CleanupOldDeletedAssets(ctx); err != nil {
		logger.Error("❌ Failed to cleanup old deleted assets", "error", err)
	} else {
		logger.Info("🧹 Successfully cleaned up old deleted assets (older than 7 days)")
	}
}
