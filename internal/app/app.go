package app

import (
	"context"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"eclat/internal/watcher"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state and core service dependencies.
// It serves as the main entry point for the Wails application runtime.
type App struct {
	ctx             context.Context
	logger          *slog.Logger
	Scanner         *scanner.Scanner
	SettingsService *settings.SettingsService
	Watcher         *watcher.Service
}

// NewApp creates a new App application struct with injected dependencies.
func NewApp(logger *slog.Logger, scanner *scanner.Scanner, settingsService *settings.SettingsService, watcher *watcher.Service) *App {
	return &App{
		logger:          logger,
		Scanner:         scanner,
		SettingsService: settingsService,
		Watcher:         watcher,
	}
}

// OnStartup is called when the application starts. The context is saved
// so we can call the runtime methods, and background services are initialized.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
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
		runtime.WindowShow(a.ctx)
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
