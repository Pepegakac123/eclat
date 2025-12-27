package app

import (
	"context"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"eclat/internal/watcher"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	logger          *slog.Logger
	Scanner         *scanner.Scanner
	SettingsService *settings.SettingsService
	Watcher         *watcher.Service
}

func NewApp(scanner *scanner.Scanner, settingsService *settings.SettingsService, watcher *watcher.Service) *App {
	return &App{
		logger:          slog.Default(),
		Scanner:         scanner,
		SettingsService: settingsService,
		Watcher:         watcher,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.Scanner.Startup(ctx)
	a.SettingsService.Startup(ctx)
	a.Watcher.Startup(ctx)
	a.logger.Info("App started")
}

// RestoreWindow przywraca i foksuje główne okno aplikacji
// Ta metoda jest publiczna (z dużej litery), więc main.go może ją wywołać
func (a *App) RestoreWindow() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
	}
}
