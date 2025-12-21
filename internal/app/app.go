package app

import (
	"context"
	"eclat/internal/scanner"
	"eclat/internal/settings"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	logger          *slog.Logger
	Scanner         *scanner.Scanner
	SettingsService *settings.SettingsService
}

func NewApp(scanner *scanner.Scanner, settingsService *settings.SettingsService) *App {
	return &App{
		logger:          slog.Default(),
		Scanner:         scanner,
		SettingsService: settingsService,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.Scanner.Startup(ctx)
	a.SettingsService.Startup(ctx)
	a.logger.Info("App started")
}

// RestoreWindow przywraca i foksuje główne okno aplikacji
// Ta metoda jest publiczna (z dużej litery), więc main.go może ją wywołać
func (a *App) RestoreWindow() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
	}
}
