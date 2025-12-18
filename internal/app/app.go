package app

import (
	"context"
	"eclat/internal/services"
	"log/slog"
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
