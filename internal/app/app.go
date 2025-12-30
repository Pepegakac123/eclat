package app

import (
	"context"
	"eclat/internal/database"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"eclat/internal/watcher"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state and core service dependencies.
// It serves as the main entry point for the Wails application runtime.
type App struct {
	db                 database.Querier
	ctx                context.Context
	logger             *slog.Logger
	AssetService       *AssetService
	MaterialSetService *MaterialSetService
	Scanner            *scanner.Scanner
	SettingsService    *settings.SettingsService
	Watcher            *watcher.Service
}

// NewApp creates a new App application struct with injected dependencies.
func NewApp(db database.Querier, logger *slog.Logger, assetService *AssetService, materialSetService *MaterialSetService, scanner *scanner.Scanner, settingsService *settings.SettingsService, watcher *watcher.Service) *App {
	return &App{
		db:                 db,
		logger:             logger,
		AssetService:       assetService,
		MaterialSetService: materialSetService,
		Scanner:            scanner,
		SettingsService:    settingsService,
		Watcher:            watcher,
	}
}

// OnStartup is called when the application starts. The context is saved
// so we can call the runtime methods, and background services are initialized.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.AssetService.Startup(ctx)
	a.MaterialSetService.Startup(ctx)
	a.Scanner.Startup(ctx)
	a.SettingsService.Startup(ctx)
	a.Watcher.Startup(ctx)

	// Start listening for watcher events to trigger scanner updates
	go a.Scanner.ListenToWatcher(a.Watcher.Events)

	a.logger.Info("App started")
}

// RestoreWindow restores and focuses the main application window.
// This is exposed publicly to allow invocation from the single instance lock mechanism in main.go.
func (a *App) RestoreWindow() {
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
	}
}

// Shutdown is called at application termination.
// It performs cleanup tasks such as stopping the watcher service.
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("🛑 App shutting down...")

	if a.Watcher != nil {
		a.Watcher.Shutdown()
	}
}

// OpenInExplorer opens the file explorer and selects the file at the given path.
func (a *App) OpenInExplorer(path string) error {
	a.logger.Info("Opening in explorer", "path", path)
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "linux":
		// Try dbus-send first for selecting file (Nautilus, Dolphin, etc.)
		// This is experimental, fallback to opening parent dir
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// OpenInDefaultApp opens the file using the system's default application.
func (a *App) OpenInDefaultApp(path string) error {
	a.logger.Info("Opening in default app", "path", path)
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
