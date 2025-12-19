package app

import (
	"context"
	"eclat/internal/services"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	logger          *slog.Logger
	Scanner         *services.Scanner
	SettingsService *services.SettingsService
}

func NewApp(scanner *services.Scanner, settingsService *services.SettingsService) *App {
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
