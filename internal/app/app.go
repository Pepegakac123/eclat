package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"eclat/internal/services"
	"log/slog"
	"os"
	"path/filepath"
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
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic("Could not get user cache directory: " + err.Error())
	}
	appCachePath := filepath.Join(userCacheDir, "eclat", "thumbnails")
	if err := os.MkdirAll(appCachePath, 0755); err != nil {
		panic("Could not create cache directory: " + err.Error())
	}
	thumbGenerator := services.NewThumbnailGenerator(appCachePath, slog.Default())
	return &App{
		db:      queries,
		logger:  slog.Default(),
		scanner: services.NewScanner(queries, thumbGenerator),
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}
