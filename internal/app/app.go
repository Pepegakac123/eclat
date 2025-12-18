package app

import (
	"context"
	"eclat/internal/services"
	"log/slog"
)

// App struct
type App struct {
	ctx     context.Context
	logger  *slog.Logger
	Scanner *services.Scanner
}

func NewApp(scanner *services.Scanner) *App {
	return &App{
		logger:  slog.Default(),
		Scanner: scanner,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.Scanner.Startup(ctx)
	a.logger.Info("App started")
}
