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

	"eclat/internal/app"
	"eclat/internal/config"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"eclat/internal/watcher"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // SQLite driver
)

// Dependencies holds all the initialized services and resources required by the application.
type Dependencies struct {
	DB                 *sql.DB
	Logger             *slog.Logger
	App                *app.App
	AssetService       *app.AssetService
	MaterialSetService *app.MaterialSetService
	ScannerService     *scanner.Scanner
	SettingsService    *settings.SettingsService
	WatcherService     *watcher.Service
	ThumbnailsDir      string
}

// Initialize performs the startup sequence: configuring directories, logger, database, and services.
func Initialize(migrations embed.FS) (*Dependencies, error) {
	// 1. Setup Directories
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
	logFilePath := filepath.Join(logsFolder, "app.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	// Note: logFile is not closed here; the OS will close it on exit, which is acceptable for the main app logger.

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	programLogger := slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
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
	assetService := app.NewAssetService(queries, db, programLogger)
	materialSetService := app.NewMaterialSetService(queries, programLogger)

	myApp := app.NewApp(queries, programLogger, assetService, materialSetService, scannerService, settingsService, watcherService)

	return &Dependencies{
		DB:                 db,
		Logger:             programLogger,
		App:                myApp,
		AssetService:       assetService,
		MaterialSetService: materialSetService,
		ScannerService:     scannerService,
		SettingsService:    settingsService,
		WatcherService:     watcherService,
		ThumbnailsDir:      thumbsFolder,
	}, nil
}
