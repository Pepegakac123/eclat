package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"log/slog"
)

// App struct
type App struct {
	ctx    context.Context
	db     *database.Queries // Dostęp do metod sqlc
	logger *slog.Logger
}

// NewApp tworzy nową instancję App z wstrzykniętymi zależnościami
func NewApp(db *sql.DB) *App {
	return &App{
		db:     database.New(db), // Wrapujemy połączenie SQL w sqlc
		logger: slog.Default(),
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}
