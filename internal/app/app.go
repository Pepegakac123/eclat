package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"eclat/internal/services"
	"log/slog"
)

// App struct
type App struct {
	ctx     context.Context
	db      *database.Queries // Dostęp do metod sqlc
	logger  *slog.Logger
	scanner *services.Scanner
}

// NewApp Creates a new instance of App with injected dependencies
func NewApp(db *sql.DB) *App {
	queries := database.New(db)
	return &App{
		db:      queries,
		logger:  slog.Default(),
		scanner: services.NewScanner(queries),
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}
